package reward

import (
	"context"
	"time"

	"github.com/google/uuid"

	"hexagonal/internal/model"
	"hexagonal/internal/repository"
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
			Tier:            model.TierForPoints(customer.Points),
			RemainingPoints: customer.Points,
		}
		return nil
	})
	if err != nil {
		return RedeemResult{}, err
	}
	return result, nil
}
