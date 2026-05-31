package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	customermodel "hexagonal-modular/internal/customer/model"
	"hexagonal-modular/internal/reward/model"
	"hexagonal-modular/internal/shared/mail"
	sharedmodel "hexagonal-modular/internal/shared/model"
	"hexagonal-modular/internal/shared/repository"
)

type RedeemResult struct {
	Reward          model.Reward
	Customer        customermodel.Customer
	Tier            sharedmodel.Tier
	RemainingPoints int
}

func (s *Service) Redeem(ctx context.Context, customerID uuid.UUID, rewardType model.RewardType) (RedeemResult, error) {
	cost, ok := model.RewardCosts[rewardType]
	if !ok {
		return RedeemResult{}, model.ErrUnknownReward
	}

	var result RedeemResult

	err := s.tx.RunTx(ctx, func(repos repository.TxRepos) error {
		customer, err := repos.Customers.FindByID(ctx, customerID)
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
		if err := repos.Rewards.Save(ctx, &reward); err != nil {
			return err
		}

		customer.Points -= cost
		if err := repos.Customers.UpdatePoints(ctx, customer.ID, customer.Points); err != nil {
			return err
		}

		result = RedeemResult{
			Reward:          reward,
			Customer:        customer,
			Tier:            sharedmodel.TierForPoints(customer.Points),
			RemainingPoints: customer.Points,
		}
		return nil
	})
	if err != nil {
		return RedeemResult{}, err
	}

	// Side-effect: redemption confirmation email. Log-and-continue on failure.
	subject, body := mail.RewardRedeemed(result.Customer, rewardType, cost, result.RemainingPoints)
	if err := s.mailer.Send(ctx, result.Customer.Email, subject, body); err != nil {
		slog.ErrorContext(ctx, "send reward confirmation", "err", err, "customer_id", result.Customer.ID)
	}

	return result, nil
}
