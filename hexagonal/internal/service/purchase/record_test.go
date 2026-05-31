package purchase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"hexagonal/internal/model"
	purchasesvc "hexagonal/internal/service/purchase"
	"hexagonal/internal/validate"
)

func TestService_Record(t *testing.T) {
	t.Run("bronze customer earns 1x and saves inside a transaction", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		customer := model.Customer{ID: customerID, Points: 100} // Bronze

		mocks.Customers.EXPECT().FindByID(ctx, customerID).Return(customer, nil)
		mocks.expectTxRun(ctx)
		mocks.Purchases.EXPECT().Save(ctx, mock.AnythingOfType("*model.Purchase")).Return(nil)
		mocks.Customers.EXPECT().UpdatePoints(ctx, customerID, 150).Return(nil)

		out, err := svc.Record(ctx, purchasesvc.RecordInput{
			CustomerID:  customerID,
			AmountCents: 5000, // $50 * 1.0 = 50 points
		})
		require.NoError(t, err)
		assert.Equal(t, 50, out.Purchase.PointsEarned)
		assert.Equal(t, "Bronze", out.PreviousTier.Name)
		assert.Equal(t, "Bronze", out.NewTier.Name)
		assert.False(t, out.Duplicate)
	})

	t.Run("silver multiplier kicks in and tier upgrade is reported", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		customer := model.Customer{ID: customerID, Points: 450} // Bronze, about to cross to Silver

		mocks.Customers.EXPECT().FindByID(ctx, customerID).Return(customer, nil)
		mocks.expectTxRun(ctx)
		mocks.Purchases.EXPECT().Save(ctx, mock.AnythingOfType("*model.Purchase")).Return(nil)
		mocks.Customers.EXPECT().UpdatePoints(ctx, customerID, 550).Return(nil) // 450 + 100

		out, err := svc.Record(ctx, purchasesvc.RecordInput{
			CustomerID:  customerID,
			AmountCents: 10000, // $100 * 1.0 (still Bronze at purchase time) = 100 points
		})
		require.NoError(t, err)
		assert.Equal(t, 100, out.Purchase.PointsEarned)
		assert.Equal(t, "Bronze", out.PreviousTier.Name)
		assert.Equal(t, "Silver", out.NewTier.Name)
	})

	t.Run("invalid amount rejects before any repo call", func(t *testing.T) {
		svc, _ := newService(t)
		_, err := svc.Record(context.Background(), purchasesvc.RecordInput{
			CustomerID:  uuid.New(),
			AmountCents: -1,
		})
		require.ErrorIs(t, err, validate.ErrInvalidInput)
	})

	t.Run("invalid external payment id rejects before any repo call", func(t *testing.T) {
		svc, _ := newService(t)
		bad := "bad id with spaces"
		_, err := svc.Record(context.Background(), purchasesvc.RecordInput{
			CustomerID:        uuid.New(),
			AmountCents:       100,
			ExternalPaymentID: &bad,
		})
		require.ErrorIs(t, err, validate.ErrInvalidInput)
	})

	t.Run("customer not found propagates ErrCustomerNotFound", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		mocks.Customers.EXPECT().FindByID(ctx, customerID).Return(model.Customer{}, model.ErrCustomerNotFound)

		_, err := svc.Record(ctx, purchasesvc.RecordInput{
			CustomerID:  customerID,
			AmountCents: 5000,
		})
		require.ErrorIs(t, err, model.ErrCustomerNotFound)
	})

	t.Run("idempotency returns existing purchase when external_payment_id matches", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		extID := "pi_abc123"
		existing := model.Purchase{
			ID:                uuid.New(),
			CustomerID:        uuid.New(),
			AmountCents:       5000,
			PointsEarned:      50,
			ExternalPaymentID: &extID,
		}
		owner := model.Customer{ID: existing.CustomerID, Points: 50}

		mocks.Purchases.EXPECT().FindByExternalPaymentID(ctx, extID).Return(existing, nil)
		mocks.Customers.EXPECT().FindByID(ctx, existing.CustomerID).Return(owner, nil)

		out, err := svc.Record(ctx, purchasesvc.RecordInput{
			CustomerID:        existing.CustomerID,
			AmountCents:       5000,
			ExternalPaymentID: &extID,
		})
		require.NoError(t, err)
		assert.True(t, out.Duplicate)
		assert.Equal(t, existing.ID, out.Purchase.ID)
	})

	t.Run("race duplicate during transaction is recovered as idempotent success", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		extID := "pi_race"
		customerID := uuid.New()
		existing := model.Purchase{
			ID:                uuid.New(),
			CustomerID:        customerID,
			AmountCents:       5000,
			PointsEarned:      50,
			ExternalPaymentID: &extID,
		}

		// First lookup (idempotency precheck) returns not found.
		mocks.Purchases.EXPECT().FindByExternalPaymentID(ctx, extID).
			Return(model.Purchase{}, model.ErrPurchaseNotFound).Once()
		mocks.Customers.EXPECT().FindByID(ctx, customerID).
			Return(model.Customer{ID: customerID, Points: 0}, nil).Once()
		mocks.expectTxRun(ctx)
		// The Save inside the tx loses the race.
		mocks.Purchases.EXPECT().Save(ctx, mock.AnythingOfType("*model.Purchase")).
			Return(model.ErrDuplicatePurchase)
		// Recovery: re-fetch existing purchase + customer.
		mocks.Purchases.EXPECT().FindByExternalPaymentID(ctx, extID).Return(existing, nil).Once()
		mocks.Customers.EXPECT().FindByID(ctx, customerID).
			Return(model.Customer{ID: customerID, Points: 50}, nil).Once()

		out, err := svc.Record(ctx, purchasesvc.RecordInput{
			CustomerID:        customerID,
			AmountCents:       5000,
			ExternalPaymentID: &extID,
		})
		require.NoError(t, err)
		assert.True(t, out.Duplicate)
		assert.Equal(t, existing.ID, out.Purchase.ID)
	})

	t.Run("OccurredAt is respected when provided", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		customerID := uuid.New()
		occurred := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)

		mocks.Customers.EXPECT().FindByID(ctx, customerID).Return(model.Customer{ID: customerID}, nil)
		mocks.expectTxRun(ctx)
		var saved *model.Purchase
		mocks.Purchases.EXPECT().Save(ctx, mock.AnythingOfType("*model.Purchase")).
			RunAndReturn(func(_ context.Context, p *model.Purchase) error {
				saved = p
				return nil
			})
		mocks.Customers.EXPECT().UpdatePoints(ctx, customerID, mock.Anything).Return(nil)

		_, err := svc.Record(ctx, purchasesvc.RecordInput{
			CustomerID:  customerID,
			AmountCents: 1000,
			OccurredAt:  occurred,
		})
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Equal(t, occurred, saved.CreatedAt)
	})
}
