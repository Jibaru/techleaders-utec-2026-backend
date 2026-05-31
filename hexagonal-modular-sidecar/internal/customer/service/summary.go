package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	customermodel "hexagonal-modular-sidecar/internal/customer/model"
	"hexagonal-modular-sidecar/internal/shared/model"
)

type Summary struct {
	Customer          customermodel.Customer
	CurrentTier       model.Tier
	NextTier          *model.Tier
	PointsToNextTier  int
	TotalSpentCents   int64
	TotalPointsEarned int
	TotalPointsSpent  int
	PurchaseCount     int
	AvgPurchaseCents  int64
	RewardCount       int
	LastPurchaseAt    *time.Time
}

func (s *Service) GetSummary(ctx context.Context, id uuid.UUID) (Summary, error) {
	customer, err := s.customers.FindByID(ctx, id)
	if err != nil {
		return Summary{}, err
	}

	purchaseAgg, err := s.purchases.AggregateForCustomer(ctx, id)
	if err != nil {
		return Summary{}, err
	}
	rewardAgg, err := s.rewards.AggregateForCustomer(ctx, id)
	if err != nil {
		return Summary{}, err
	}

	current := model.TierForPoints(customer.Points)
	next := model.NextTierForPoints(customer.Points)
	pointsToNext := 0
	if next != nil {
		pointsToNext = next.MinPoints - customer.Points
		if pointsToNext < 0 {
			pointsToNext = 0
		}
	}

	var avg int64
	if purchaseAgg.Count > 0 {
		avg = purchaseAgg.TotalSpent / purchaseAgg.Count
	}

	return Summary{
		Customer:          customer,
		CurrentTier:       current,
		NextTier:          next,
		PointsToNextTier:  pointsToNext,
		TotalSpentCents:   purchaseAgg.TotalSpent,
		TotalPointsEarned: int(purchaseAgg.TotalEarned),
		TotalPointsSpent:  int(rewardAgg.TotalSpent),
		PurchaseCount:     int(purchaseAgg.Count),
		AvgPurchaseCents:  avg,
		RewardCount:       int(rewardAgg.Count),
		LastPurchaseAt:    purchaseAgg.LastPurchaseAt,
	}, nil
}
