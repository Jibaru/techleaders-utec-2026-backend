package cached_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"hexagonal-modular-sidecar/internal/customer/model"
	customermock "hexagonal-modular-sidecar/internal/customer/repository/mock"
	"hexagonal-modular-sidecar/internal/customer/repository/cached"
	"hexagonal-modular-sidecar/internal/shared/cache"
	cachemock "hexagonal-modular-sidecar/internal/shared/cache/mock"
)

// TestRepository_FindByID covers the cache-aside read path.
func TestRepository_FindByID(t *testing.T) {
	t.Run("cache hit returns customer without touching the inner repo", func(t *testing.T) {
		ctx := context.Background()
		id := uuid.New()
		want := model.Customer{ID: id, Name: "Ana", Email: "ana@example.com", Points: 120}
		raw, err := json.Marshal(want)
		require.NoError(t, err)

		inner := customermock.NewRepository(t)
		cch := cachemock.NewCache(t)
		cch.EXPECT().Get(ctx, "customer:"+id.String()).Return(raw, nil)

		r := cached.New(inner, cch)
		got, err := r.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, want, got)
		// inner.AssertExpectations runs via t.Cleanup — if FindByID had been
		// called we'd see "no expectation" panic.
	})

	t.Run("cache miss falls back to inner repo and populates cache", func(t *testing.T) {
		ctx := context.Background()
		id := uuid.New()
		want := model.Customer{ID: id, Name: "Bruno", Email: "bruno@example.com", Points: 750}

		inner := customermock.NewRepository(t)
		cch := cachemock.NewCache(t)
		cch.EXPECT().Get(ctx, "customer:"+id.String()).Return(nil, cache.ErrMiss)
		inner.EXPECT().FindByID(ctx, id).Return(want, nil)
		cch.EXPECT().Set(ctx, "customer:"+id.String(),
			mock.AnythingOfType("[]uint8"), mock.AnythingOfType("time.Duration")).Return(nil)

		r := cached.New(inner, cch)
		got, err := r.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("cache transport error is logged-and-ignored; falls back to inner", func(t *testing.T) {
		ctx := context.Background()
		id := uuid.New()
		want := model.Customer{ID: id, Name: "Carla"}

		inner := customermock.NewRepository(t)
		cch := cachemock.NewCache(t)
		// Redis hiccup: returns a non-miss error
		cch.EXPECT().Get(ctx, "customer:"+id.String()).Return(nil, errors.New("redis: dial timeout"))
		inner.EXPECT().FindByID(ctx, id).Return(want, nil)
		cch.EXPECT().Set(ctx, "customer:"+id.String(),
			mock.AnythingOfType("[]uint8"), mock.AnythingOfType("time.Duration")).Return(nil)

		r := cached.New(inner, cch)
		got, err := r.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("corrupt cache payload triggers fallback to inner", func(t *testing.T) {
		ctx := context.Background()
		id := uuid.New()
		want := model.Customer{ID: id, Name: "Diego"}

		inner := customermock.NewRepository(t)
		cch := cachemock.NewCache(t)
		cch.EXPECT().Get(ctx, "customer:"+id.String()).Return([]byte("not-json"), nil)
		inner.EXPECT().FindByID(ctx, id).Return(want, nil)
		cch.EXPECT().Set(ctx, "customer:"+id.String(),
			mock.AnythingOfType("[]uint8"), mock.AnythingOfType("time.Duration")).Return(nil)

		r := cached.New(inner, cch)
		got, err := r.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("inner repo error is propagated; cache is not populated", func(t *testing.T) {
		ctx := context.Background()
		id := uuid.New()

		inner := customermock.NewRepository(t)
		cch := cachemock.NewCache(t)
		cch.EXPECT().Get(ctx, "customer:"+id.String()).Return(nil, cache.ErrMiss)
		inner.EXPECT().FindByID(ctx, id).Return(model.Customer{}, model.ErrCustomerNotFound)
		// No cch.EXPECT().Set — must not be called when inner errors.

		r := cached.New(inner, cch)
		_, err := r.FindByID(ctx, id)
		require.ErrorIs(t, err, model.ErrCustomerNotFound)
	})
}

func TestRepository_UpdatePoints(t *testing.T) {
	t.Run("write-through invalidates the cache after a successful update", func(t *testing.T) {
		ctx := context.Background()
		id := uuid.New()

		inner := customermock.NewRepository(t)
		cch := cachemock.NewCache(t)
		inner.EXPECT().UpdatePoints(ctx, id, 500).Return(nil)
		cch.EXPECT().Delete(ctx, "customer:"+id.String()).Return(nil)

		r := cached.New(inner, cch)
		require.NoError(t, r.UpdatePoints(ctx, id, 500))
	})

	t.Run("inner failure prevents cache invalidation", func(t *testing.T) {
		ctx := context.Background()
		id := uuid.New()

		inner := customermock.NewRepository(t)
		cch := cachemock.NewCache(t)
		inner.EXPECT().UpdatePoints(ctx, id, 500).Return(errors.New("db down"))
		// No cch.EXPECT().Delete — must not be called.

		r := cached.New(inner, cch)
		err := r.UpdatePoints(ctx, id, 500)
		require.Error(t, err)
	})
}

func TestRepository_Update(t *testing.T) {
	t.Run("write-through invalidates the cache and returns the updated customer", func(t *testing.T) {
		ctx := context.Background()
		id := uuid.New()
		updated := model.Customer{ID: id, Name: "New Name"}

		inner := customermock.NewRepository(t)
		cch := cachemock.NewCache(t)
		updates := map[string]any{"name": "New Name"}
		inner.EXPECT().Update(ctx, id, updates).Return(updated, nil)
		cch.EXPECT().Delete(ctx, "customer:"+id.String()).Return(nil)

		r := cached.New(inner, cch)
		got, err := r.Update(ctx, id, updates)
		require.NoError(t, err)
		assert.Equal(t, updated, got)
	})
}
