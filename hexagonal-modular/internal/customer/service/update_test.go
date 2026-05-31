package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hexagonal-modular/internal/customer/model"
	customersvc "hexagonal-modular/internal/customer/service"
	"hexagonal-modular/internal/shared/validate"
)

func ptr(s string) *string { return &s }

func TestService_Update(t *testing.T) {
	t.Run("happy path with both fields applies trimmed values", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		id := uuid.New()
		updated := model.Customer{ID: id, Name: "Ana Reyes", Email: "ana.reyes@example.com", Points: 50}

		mocks.Customers.EXPECT().FindByID(ctx, id).Return(model.Customer{ID: id}, nil)
		mocks.Customers.EXPECT().Update(ctx, id, map[string]any{
			"name":  "Ana Reyes",
			"email": "ana.reyes@example.com",
		}).Return(updated, nil)

		got, err := svc.Update(ctx, id, customersvc.UpdateInput{
			Name:  ptr("  Ana Reyes  "),
			Email: ptr("  Ana.Reyes@Example.COM  "),
		})
		require.NoError(t, err)
		assert.Equal(t, updated, got)
	})

	t.Run("happy path with only name applies just that field", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		id := uuid.New()
		updated := model.Customer{ID: id, Name: "Renamed"}

		mocks.Customers.EXPECT().FindByID(ctx, id).Return(model.Customer{ID: id}, nil)
		mocks.Customers.EXPECT().Update(ctx, id, map[string]any{"name": "Renamed"}).
			Return(updated, nil)

		got, err := svc.Update(ctx, id, customersvc.UpdateInput{Name: ptr("Renamed")})
		require.NoError(t, err)
		assert.Equal(t, updated, got)
	})

	t.Run("invalid name rejects before touching repo", func(t *testing.T) {
		svc, _ := newService(t)
		_, err := svc.Update(context.Background(), uuid.New(), customersvc.UpdateInput{Name: ptr("12345")})
		require.ErrorIs(t, err, validate.ErrInvalidInput)
	})

	t.Run("invalid email rejects before touching repo", func(t *testing.T) {
		svc, _ := newService(t)
		_, err := svc.Update(context.Background(), uuid.New(), customersvc.UpdateInput{Email: ptr("nope")})
		require.ErrorIs(t, err, validate.ErrInvalidInput)
	})

	t.Run("empty input returns validation error", func(t *testing.T) {
		svc, _ := newService(t)
		_, err := svc.Update(context.Background(), uuid.New(), customersvc.UpdateInput{})
		require.ErrorIs(t, err, validate.ErrInvalidInput)
	})

	t.Run("customer not found propagates the error", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		id := uuid.New()
		mocks.Customers.EXPECT().FindByID(ctx, id).Return(model.Customer{}, model.ErrCustomerNotFound)

		_, err := svc.Update(ctx, id, customersvc.UpdateInput{Name: ptr("Ana Reyes")})
		require.ErrorIs(t, err, model.ErrCustomerNotFound)
	})

	t.Run("email collision propagates ErrEmailAlreadyExists", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		id := uuid.New()
		mocks.Customers.EXPECT().FindByID(ctx, id).Return(model.Customer{ID: id}, nil)
		mocks.Customers.EXPECT().Update(ctx, id, map[string]any{"email": "taken@example.com"}).
			Return(model.Customer{}, model.ErrEmailAlreadyExists)

		_, err := svc.Update(ctx, id, customersvc.UpdateInput{Email: ptr("taken@example.com")})
		require.ErrorIs(t, err, model.ErrEmailAlreadyExists)
	})
}
