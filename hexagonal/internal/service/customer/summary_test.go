package customer_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hexagonal/internal/model"
	purchaserepo "hexagonal/internal/repository/purchase"
	rewardrepo "hexagonal/internal/repository/reward"
)

func TestService_GetSummary(t *testing.T) {
	t.Run("happy path aggregates purchases and rewards and computes next-tier", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		id := uuid.New()
		lastPurchase := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

		mocks.Customers.EXPECT().FindByID(ctx, id).Return(model.Customer{ID: id, Points: 750}, nil)
		mocks.Purchases.EXPECT().AggregateForCustomer(ctx, id).Return(purchaserepo.Aggregate{
			TotalSpent: 60000, TotalEarned: 800, Count: 4, LastPurchaseAt: &lastPurchase,
		}, nil)
		mocks.Rewards.EXPECT().AggregateForCustomer(ctx, id).Return(rewardrepo.Aggregate{
			TotalSpent: 50, Count: 1,
		}, nil)

		s, err := svc.GetSummary(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, 750, s.Customer.Points)
		assert.Equal(t, "Silver", s.CurrentTier.Name)
		require.NotNil(t, s.NextTier)
		assert.Equal(t, "Gold", s.NextTier.Name)
		assert.Equal(t, 1250, s.PointsToNextTier) // 2000 - 750
		assert.Equal(t, int64(60000), s.TotalSpentCents)
		assert.Equal(t, 800, s.TotalPointsEarned)
		assert.Equal(t, 50, s.TotalPointsSpent)
		assert.Equal(t, 4, s.PurchaseCount)
		assert.Equal(t, int64(15000), s.AvgPurchaseCents) // 60000 / 4
		assert.Equal(t, 1, s.RewardCount)
		assert.Equal(t, &lastPurchase, s.LastPurchaseAt)
	})

	t.Run("gold customer has no next tier", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		id := uuid.New()

		mocks.Customers.EXPECT().FindByID(ctx, id).Return(model.Customer{ID: id, Points: 2500}, nil)
		mocks.Purchases.EXPECT().AggregateForCustomer(ctx, id).Return(purchaserepo.Aggregate{}, nil)
		mocks.Rewards.EXPECT().AggregateForCustomer(ctx, id).Return(rewardrepo.Aggregate{}, nil)

		s, err := svc.GetSummary(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "Gold", s.CurrentTier.Name)
		assert.Nil(t, s.NextTier)
		assert.Equal(t, 0, s.PointsToNextTier)
	})

	t.Run("customer not found propagates the error", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		id := uuid.New()
		mocks.Customers.EXPECT().FindByID(ctx, id).Return(model.Customer{}, model.ErrCustomerNotFound)

		_, err := svc.GetSummary(ctx, id)
		require.ErrorIs(t, err, model.ErrCustomerNotFound)
	})
}
