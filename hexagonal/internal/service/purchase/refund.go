package purchase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"hexagonal/internal/model"
	"hexagonal/internal/repository"
)

type RefundResult struct {
	Purchase       model.Purchase
	Customer       model.Customer
	PointsReverted int
	PreviousTier   model.Tier
	NewTier        model.Tier
}

func (s *Service) Refund(ctx context.Context, purchaseID uuid.UUID) (RefundResult, error) {
	var result RefundResult

	err := s.tx.RunTx(ctx, func(repos repository.TxRepos) error {
		purchase, err := repos.Purchases.FindByID(ctx, purchaseID)
		if err != nil {
			return err
		}
		if purchase.RefundedAt != nil {
			return model.ErrAlreadyRefunded
		}

		customer, err := repos.Customers.FindByID(ctx, purchase.CustomerID)
		if err != nil {
			return err
		}
		if customer.Points < purchase.PointsEarned {
			return model.ErrPointsAlreadySpent
		}

		previousTier := model.TierForPoints(customer.Points)
		customer.Points -= purchase.PointsEarned
		now := time.Now().UTC()
		purchase.RefundedAt = &now

		if err := repos.Customers.UpdatePoints(ctx, customer.ID, customer.Points); err != nil {
			return err
		}
		if err := repos.Purchases.MarkRefunded(ctx, purchase.ID, now); err != nil {
			return err
		}

		result = RefundResult{
			Purchase:       purchase,
			Customer:       customer,
			PointsReverted: purchase.PointsEarned,
			PreviousTier:   previousTier,
			NewTier:        model.TierForPoints(customer.Points),
		}
		return nil
	})
	if err != nil {
		return RefundResult{}, err
	}
	return result, nil
}
