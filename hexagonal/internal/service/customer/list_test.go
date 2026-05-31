package customer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"hexagonal/internal/model"
	customerrepo "hexagonal/internal/repository/customer"
	customersvc "hexagonal/internal/service/customer"
	"hexagonal/internal/validate"
)

func TestService_List(t *testing.T) {
	t.Run("happy path applies defaults and forwards to repo", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		want := []model.Customer{
			{Name: "Ana", Points: 50},
			{Name: "Bruno", Points: 750},
		}
		mocks.Customers.EXPECT().List(ctx, customerrepo.ListFilter{
			MinPoints: nil, MaxPoints: nil, Sort: "", Limit: 20, Offset: 0,
		}).Return(want, int64(2), nil)

		out, err := svc.List(ctx, customersvc.ListInput{})
		require.NoError(t, err)
		assert.Equal(t, want, out.Customers)
		assert.Equal(t, int64(2), out.Total)
		assert.Equal(t, 1, out.Page)
		assert.Equal(t, 20, out.Limit)
		assert.Equal(t, 1, out.TotalPages)
	})

	t.Run("tier filter translates to a point range", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		mocks.Customers.EXPECT().List(ctx, mock.MatchedBy(func(f customerrepo.ListFilter) bool {
			return f.MinPoints != nil && *f.MinPoints == 500 &&
				f.MaxPoints != nil && *f.MaxPoints == 1999
		})).Return(nil, int64(0), nil)

		_, err := svc.List(ctx, customersvc.ListInput{Tier: "silver"})
		require.NoError(t, err)
	})

	t.Run("invalid sort rejects without calling repo", func(t *testing.T) {
		svc, _ := newService(t)
		_, err := svc.List(context.Background(), customersvc.ListInput{Sort: "by-vibes"})
		require.ErrorIs(t, err, validate.ErrInvalidInput)
	})

	t.Run("unknown tier rejects without calling repo", func(t *testing.T) {
		svc, _ := newService(t)
		_, err := svc.List(context.Background(), customersvc.ListInput{Tier: "platinum"})
		require.ErrorIs(t, err, validate.ErrInvalidInput)
	})

	t.Run("limit capped at 100 and pagination math works", func(t *testing.T) {
		svc, mocks := newService(t)
		ctx := context.Background()
		mocks.Customers.EXPECT().List(ctx, mock.MatchedBy(func(f customerrepo.ListFilter) bool {
			return f.Limit == 100 && f.Offset == 200
		})).Return(nil, int64(345), nil)

		out, err := svc.List(ctx, customersvc.ListInput{Page: 3, Limit: 500})
		require.NoError(t, err)
		assert.Equal(t, 100, out.Limit)
		assert.Equal(t, 4, out.TotalPages) // ceil(345/100)
	})
}
