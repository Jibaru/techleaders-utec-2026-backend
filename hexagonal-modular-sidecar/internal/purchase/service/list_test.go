package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	customermodel "hexagonal-modular-sidecar/internal/customer/model"
	"hexagonal-modular-sidecar/internal/purchase/model"
)

func TestService_List(t *testing.T) {
	t.Run("happy path returns purchases for the customer", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		want := []model.Purchase{
			{ID: uuid.New(), CustomerID: customerID, AmountCents: 1000, PointsEarned: 10},
		}

		mocks.Customers.EXPECT().ExistsByID(ctx, customerID).Return(true, nil)
		mocks.Purchases.EXPECT().ListByCustomer(ctx, customerID).Return(want, nil)

		got, err := svc.List(ctx, customerID)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("customer not found returns ErrCustomerNotFound", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		mocks.Customers.EXPECT().ExistsByID(ctx, customerID).Return(false, nil)

		_, err := svc.List(ctx, customerID)
		require.ErrorIs(t, err, customermodel.ErrCustomerNotFound)
	})
}
