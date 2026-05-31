package customer_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hexagonal/internal/model"
)

func TestService_Get(t *testing.T) {
	t.Run("happy path returns the customer from the repo", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		id := uuid.New()
		want := model.Customer{ID: id, Name: "Ana", Email: "ana@example.com", Points: 120}

		mocks.Customers.EXPECT().FindByID(ctx, id).Return(want, nil)

		got, err := svc.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("not found propagates ErrCustomerNotFound", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		id := uuid.New()

		mocks.Customers.EXPECT().FindByID(ctx, id).Return(model.Customer{}, model.ErrCustomerNotFound)

		_, err := svc.Get(ctx, id)
		require.ErrorIs(t, err, model.ErrCustomerNotFound)
	})
}
