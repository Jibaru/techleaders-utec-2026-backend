package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	customermodel "hexagonal-modular/internal/customer/model"
	"hexagonal-modular/internal/purchase/model"
)

func TestService_Refund(t *testing.T) {
	t.Run("happy path reverts points and marks the purchase refunded", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		purchaseID := uuid.New()
		purchase := model.Purchase{ID: purchaseID, CustomerID: customerID, PointsEarned: 50}
		customer := customermodel.Customer{ID: customerID, Points: 750}

		mocks.expectTxRun(ctx)
		mocks.Purchases.EXPECT().FindByID(ctx, purchaseID).Return(purchase, nil)
		mocks.Customers.EXPECT().FindByID(ctx, customerID).Return(customer, nil)
		mocks.Customers.EXPECT().UpdatePoints(ctx, customerID, 700).Return(nil)
		mocks.Purchases.EXPECT().MarkRefunded(ctx, purchaseID, mock.AnythingOfType("time.Time")).Return(nil)

		out, err := svc.Refund(ctx, purchaseID)
		require.NoError(t, err)
		assert.Equal(t, 50, out.PointsReverted)
		assert.Equal(t, "Silver", out.PreviousTier.Name) // 750 → Silver
		assert.Equal(t, "Silver", out.NewTier.Name)      // 700 → still Silver
		assert.NotNil(t, out.Purchase.RefundedAt)
	})

	t.Run("refund crossing tier boundary reports tier demotion", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		purchaseID := uuid.New()
		// Customer is at 520 (Silver). Refunding 100 points sends them to 420 (Bronze).
		purchase := model.Purchase{ID: purchaseID, CustomerID: customerID, PointsEarned: 100}
		customer := customermodel.Customer{ID: customerID, Points: 520}

		mocks.expectTxRun(ctx)
		mocks.Purchases.EXPECT().FindByID(ctx, purchaseID).Return(purchase, nil)
		mocks.Customers.EXPECT().FindByID(ctx, customerID).Return(customer, nil)
		mocks.Customers.EXPECT().UpdatePoints(ctx, customerID, 420).Return(nil)
		mocks.Purchases.EXPECT().MarkRefunded(ctx, purchaseID, mock.AnythingOfType("time.Time")).Return(nil)

		out, err := svc.Refund(ctx, purchaseID)
		require.NoError(t, err)
		assert.Equal(t, "Silver", out.PreviousTier.Name)
		assert.Equal(t, "Bronze", out.NewTier.Name)
	})

	t.Run("purchase not found propagates the error", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		purchaseID := uuid.New()
		mocks.expectTxRun(ctx)
		mocks.Purchases.EXPECT().FindByID(ctx, purchaseID).Return(model.Purchase{}, model.ErrPurchaseNotFound)

		_, err := svc.Refund(ctx, purchaseID)
		require.ErrorIs(t, err, model.ErrPurchaseNotFound)
	})

	t.Run("already refunded returns ErrAlreadyRefunded", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		purchaseID := uuid.New()
		now := mustTime("2026-01-01T00:00:00Z")
		purchase := model.Purchase{ID: purchaseID, RefundedAt: &now}

		mocks.expectTxRun(ctx)
		mocks.Purchases.EXPECT().FindByID(ctx, purchaseID).Return(purchase, nil)

		_, err := svc.Refund(ctx, purchaseID)
		require.ErrorIs(t, err, model.ErrAlreadyRefunded)
	})

	t.Run("points already spent returns ErrPointsAlreadySpent", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		purchaseID := uuid.New()
		purchase := model.Purchase{ID: purchaseID, CustomerID: customerID, PointsEarned: 100}
		customer := customermodel.Customer{ID: customerID, Points: 50} // not enough to refund

		mocks.expectTxRun(ctx)
		mocks.Purchases.EXPECT().FindByID(ctx, purchaseID).Return(purchase, nil)
		mocks.Customers.EXPECT().FindByID(ctx, customerID).Return(customer, nil)

		_, err := svc.Refund(ctx, purchaseID)
		require.ErrorIs(t, err, model.ErrPointsAlreadySpent)
	})
}
