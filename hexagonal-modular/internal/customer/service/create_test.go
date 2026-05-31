package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"hexagonal-modular/internal/customer/model"
	"hexagonal-modular/internal/shared/validate"
)

func TestService_Create(t *testing.T) {
	t.Run("happy path trims, lowercases email, saves and returns customer", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()

		var saved *model.Customer
		mocks.Customers.EXPECT().Save(ctx, mock.AnythingOfType("*model.Customer")).
			RunAndReturn(func(_ context.Context, c *model.Customer) error {
				saved = c
				return nil
			})

		c, err := svc.Create(ctx, "  Ana Torres  ", "  Ana@Example.COM  ")
		require.NoError(t, err)
		assert.Equal(t, "Ana Torres", c.Name)
		assert.Equal(t, "ana@example.com", c.Email)
		assert.Equal(t, 0, c.Points)
		assert.NotEqual(t, "", c.ID.String())
		require.NotNil(t, saved)
		assert.Equal(t, c.ID, saved.ID)
	})

	t.Run("invalid name rejects before touching repo", func(t *testing.T) {
		svc, _ := newService(t)

		_, err := svc.Create(context.Background(), "12345", "ana@example.com")
		require.ErrorIs(t, err, validate.ErrInvalidInput)
	})

	t.Run("invalid email rejects before touching repo", func(t *testing.T) {
		svc, _ := newService(t)

		_, err := svc.Create(context.Background(), "Ana Torres", "not-an-email")
		require.ErrorIs(t, err, validate.ErrInvalidInput)
	})

	t.Run("propagates email-already-exists from repo", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()

		mocks.Customers.EXPECT().Save(ctx, mock.AnythingOfType("*model.Customer")).
			Return(model.ErrEmailAlreadyExists)

		_, err := svc.Create(ctx, "Ana Torres", "ana@example.com")
		require.ErrorIs(t, err, model.ErrEmailAlreadyExists)
	})
}
