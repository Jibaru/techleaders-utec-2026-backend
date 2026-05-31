package main

import (
	"time"

	"github.com/google/uuid"

	customermodel "hexagonal-modular/internal/customer/model"
	purchasemodel "hexagonal-modular/internal/purchase/model"
	rewardmodel "hexagonal-modular/internal/reward/model"
)

// Well-known UUIDs so the seeded customers have predictable IDs you can curl
// directly without first listing them.
var (
	anaID   = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	brunoID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	carlaID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	diegoID = uuid.MustParse("44444444-4444-4444-4444-444444444444")
)

func buildFixtures() ([]customermodel.Customer, []purchasemodel.Purchase, []rewardmodel.Reward) {
	now := time.Now().UTC()

	customers := []customermodel.Customer{
		{ID: anaID, Name: "Ana Torres", Email: "ana@example.com", Points: 120, CreatedAt: now.AddDate(0, -2, 0)},
		{ID: brunoID, Name: "Bruno Diaz", Email: "bruno@example.com", Points: 750, CreatedAt: now.AddDate(0, -4, 0)},
		{ID: carlaID, Name: "Carla Reyes", Email: "carla@example.com", Points: 2100, CreatedAt: now.AddDate(0, -8, 0)},
		{ID: diegoID, Name: "Diego Vega", Email: "diego@example.com", Points: 40, CreatedAt: now.AddDate(0, 0, -3)},
	}

	purchases := []purchasemodel.Purchase{
		// Ana — Bronze, 120 pts earned total
		newPurchase(anaID, 3500, 35, now.AddDate(0, -2, 0)),
		newPurchase(anaID, 4000, 40, now.AddDate(0, -1, -15)),
		newPurchase(anaID, 4500, 45, now.AddDate(0, 0, -10)),

		// Bruno — climbed from Bronze to Silver, 825 pts earned (750 + 75 redeemed)
		newPurchase(brunoID, 15000, 150, now.AddDate(0, -4, 0)),
		newPurchase(brunoID, 20000, 200, now.AddDate(0, -3, 0)),
		newPurchase(brunoID, 22000, 275, now.AddDate(0, -2, 0)),
		newPurchase(brunoID, 16000, 200, now.AddDate(0, -1, 0)),

		// Carla — Gold customer, 2300 pts earned (2100 + 200 redeemed)
		newPurchase(carlaID, 30000, 300, now.AddDate(0, -8, 0)),
		newPurchase(carlaID, 40000, 500, now.AddDate(0, -6, 0)),
		newPurchase(carlaID, 50000, 625, now.AddDate(0, -4, 0)),
		newPurchase(carlaID, 40000, 500, now.AddDate(0, -2, 0)),
		newPurchase(carlaID, 25000, 375, now.AddDate(0, -1, 0)),

		// Diego — fresh signup, 40 pts
		newPurchase(diegoID, 4000, 40, now.AddDate(0, 0, -3)),
	}

	rewards := []rewardmodel.Reward{
		// Bruno redeemed a free pastry once
		newReward(brunoID, rewardmodel.RewardFreePastry, 75, now.AddDate(0, -1, -15)),
		// Carla redeemed two free drinks
		newReward(carlaID, rewardmodel.RewardFreeDrink, 100, now.AddDate(0, -3, 0)),
		newReward(carlaID, rewardmodel.RewardFreeDrink, 100, now.AddDate(0, -1, 0)),
	}

	return customers, purchases, rewards
}

func newPurchase(customerID uuid.UUID, amountCents int64, points int, at time.Time) purchasemodel.Purchase {
	return purchasemodel.Purchase{
		ID:           uuid.New(),
		CustomerID:   customerID,
		AmountCents:  amountCents,
		PointsEarned: points,
		CreatedAt:    at,
	}
}

func newReward(customerID uuid.UUID, t rewardmodel.RewardType, cost int, at time.Time) rewardmodel.Reward {
	return rewardmodel.Reward{
		ID:          uuid.New(),
		CustomerID:  customerID,
		Type:        t,
		PointsSpent: cost,
		CreatedAt:   at,
	}
}
