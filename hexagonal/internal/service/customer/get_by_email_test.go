package customer_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hexagonal/internal/model"
	"hexagonal/internal/validate"
)

func TestService_GetByEmail(t *testing.T) {
	t.Run("happy path lowercases and trims before lookup", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		want := model.Customer{ID: uuid.New(), Name: "Ana", Email: "ana@example.com"}

		mocks.Customers.EXPECT().FindByEmail(ctx, "ana@example.com").Return(want, nil)

		got, err := svc.GetByEmail(ctx, "  Ana@Example.COM  ")
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("invalid email rejects before touching repo", func(t *testing.T) {
		svc, _ := newService(t)

		_, err := svc.GetByEmail(context.Background(), "not-an-email")
		require.ErrorIs(t, err, validate.ErrInvalidInput)
	})

	t.Run("not found propagates ErrCustomerNotFound", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()

		mocks.Customers.EXPECT().FindByEmail(ctx, "missing@example.com").
			Return(model.Customer{}, model.ErrCustomerNotFound)

		_, err := svc.GetByEmail(ctx, "missing@example.com")
		require.ErrorIs(t, err, model.ErrCustomerNotFound)
	})
}
