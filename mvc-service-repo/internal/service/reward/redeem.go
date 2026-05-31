package reward

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mvc-service-repo/internal/mail"
	"mvc-service-repo/internal/model"
)

type RedeemResult struct {
	Reward          model.Reward
	Customer        model.Customer
	Tier            model.Tier
	RemainingPoints int
}

func (s *Service) Redeem(ctx context.Context, customerID uuid.UUID, rewardType model.RewardType) (RedeemResult, error) {
	cost, ok := model.RewardCosts[rewardType]
	if !ok {
		return RedeemResult{}, model.ErrUnknownReward
	}

	var result RedeemResult

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		customers := s.customers.WithTx(tx)
		rewards := s.rewards.WithTx(tx)

		customer, err := customers.FindByID(ctx, customerID)
		if err != nil {
			return err
		}
		if customer.Points < cost {
			return model.ErrInsufficientPoints
		}

		reward := model.Reward{
			ID:          uuid.New(),
			CustomerID:  customer.ID,
			Type:        rewardType,
			PointsSpent: cost,
			CreatedAt:   time.Now().UTC(),
		}
		if err := rewards.Save(ctx, &reward); err != nil {
			return err
		}

		customer.Points -= cost
		if err := customers.UpdatePoints(ctx, customer.ID, customer.Points); err != nil {
			return err
		}

		result = RedeemResult{
			Reward:          reward,
			Customer:        customer,
			Tier:            model.TierForPoints(customer.Points),
			RemainingPoints: customer.Points,
		}
		return nil
	})
	if err != nil {
		return RedeemResult{}, err
	}

	// Send the redemption confirmation. Log-and-continue on failure.
	subject, body := mail.RewardRedeemed(result.Customer, rewardType, cost, result.RemainingPoints)
	if err := s.mailer.Send(result.Customer.Email, subject, body); err != nil {
		slog.Error("send reward confirmation", "err", err, "customer_id", result.Customer.ID)
	}

	return result, nil
}
