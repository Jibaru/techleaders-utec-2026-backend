package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	customermodel "hexagonal-modular-sidecar/internal/customer/model"
	"hexagonal-modular-sidecar/internal/reward/model"
)

func TestService_Redeem(t *testing.T) {
	t.Run("happy path deducts the cost, saves atomically, and sends confirmation", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		customer := customermodel.Customer{ID: customerID, Email: "ana@example.com", Points: 200}

		mocks.expectTxRun(ctx)
		mocks.Customers.EXPECT().FindByID(ctx, customerID).Return(customer, nil)
		mocks.Rewards.EXPECT().Save(ctx, mock.AnythingOfType("*model.Reward")).Return(nil)
		mocks.Customers.EXPECT().UpdatePoints(ctx, customerID, 100).Return(nil) // 200 - 100 (free_drink)
		mocks.Mailer.EXPECT().Send(ctx, "ana@example.com", "Your reward is ready",
			mock.AnythingOfType("string")).Return(nil).Once()

		out, err := svc.Redeem(ctx, customerID, model.RewardFreeDrink)
		require.NoError(t, err)
		assert.Equal(t, model.RewardFreeDrink, out.Reward.Type)
		assert.Equal(t, 100, out.Reward.PointsSpent)
		assert.Equal(t, 100, out.RemainingPoints)
		assert.Equal(t, "Bronze", out.Tier.Name)
	})

	t.Run("email send failure does not fail the redemption", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		customer := customermodel.Customer{ID: customerID, Email: "ana@example.com", Points: 200}

		mocks.expectTxRun(ctx)
		mocks.Customers.EXPECT().FindByID(ctx, customerID).Return(customer, nil)
		mocks.Rewards.EXPECT().Save(ctx, mock.AnythingOfType("*model.Reward")).Return(nil)
		mocks.Customers.EXPECT().UpdatePoints(ctx, customerID, 100).Return(nil)
		mocks.Mailer.EXPECT().Send(ctx, "ana@example.com",
			mock.AnythingOfType("string"), mock.AnythingOfType("string")).
			Return(errors.New("smtp: connection refused"))

		_, err := svc.Redeem(ctx, customerID, model.RewardFreeDrink)
		require.NoError(t, err, "redemption should commit even if email send fails")
	})

	t.Run("unknown reward type returns ErrUnknownReward before any tx", func(t *testing.T) {
		svc, _ := newService(t)
		_, err := svc.Redeem(context.Background(), uuid.New(), model.RewardType("free_unicorn"))
		require.ErrorIs(t, err, model.ErrUnknownReward)
	})

	t.Run("customer not found inside tx propagates the error", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		mocks.expectTxRun(ctx)
		mocks.Customers.EXPECT().FindByID(ctx, customerID).Return(customermodel.Customer{}, customermodel.ErrCustomerNotFound)

		_, err := svc.Redeem(ctx, customerID, model.RewardFreePastry)
		require.ErrorIs(t, err, customermodel.ErrCustomerNotFound)
	})

	t.Run("insufficient points returns ErrInsufficientPoints", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		customer := customermodel.Customer{ID: customerID, Points: 10}

		mocks.expectTxRun(ctx)
		mocks.Customers.EXPECT().FindByID(ctx, customerID).Return(customer, nil)

		_, err := svc.Redeem(ctx, customerID, model.RewardFreeDrink) // 100 pts > 10 pts
		require.ErrorIs(t, err, model.ErrInsufficientPoints)
	})
}
