package service

import (
	"context"
	"fmt"
	"strings"

	customermodel "hexagonal-modular-sidecar/internal/customer/model"
	customerrepo "hexagonal-modular-sidecar/internal/customer/repository"
	"hexagonal-modular-sidecar/internal/shared/model"
	"hexagonal-modular-sidecar/internal/shared/validate"
)

// ListInput is the business-level filter spec.
type ListInput struct {
	Tier      string // "", "bronze", "silver", "gold"
	MinPoints *int
	MaxPoints *int
	Page      int
	Limit     int
	Sort      string // "", "created_at_desc", "created_at_asc", "points_desc", "points_asc"
}

type ListOutput struct {
	Customers  []customermodel.Customer
	Total      int64
	Page       int
	Limit      int
	TotalPages int
}

func (s *Service) List(ctx context.Context, in ListInput) (ListOutput, error) {
	if in.Page < 1 {
		in.Page = 1
	}
	if in.Limit < 1 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}
	if in.MinPoints != nil && *in.MinPoints < 0 {
		return ListOutput{}, fmt.Errorf("%w: min_points cannot be negative", validate.ErrInvalidInput)
	}
	if in.MaxPoints != nil && *in.MaxPoints < 0 {
		return ListOutput{}, fmt.Errorf("%w: max_points cannot be negative", validate.ErrInvalidInput)
	}
	if !isValidSort(in.Sort) {
		return ListOutput{}, fmt.Errorf("%w: unknown sort %q", validate.ErrInvalidInput, in.Sort)
	}

	min, max := in.MinPoints, in.MaxPoints
	if tier := strings.ToLower(strings.TrimSpace(in.Tier)); tier != "" {
		tierMin, tierMax, ok := pointsRangeForTier(tier)
		if !ok {
			return ListOutput{}, fmt.Errorf("%w: unknown tier %q", validate.ErrInvalidInput, in.Tier)
		}
		if min == nil || tierMin > *min {
			min = &tierMin
		}
		if tierMax >= 0 && (max == nil || tierMax < *max) {
			max = &tierMax
		}
	}

	customers, total, err := s.customers.List(ctx, customerrepo.ListFilter{
		MinPoints: min,
		MaxPoints: max,
		Sort:      in.Sort,
		Limit:     in.Limit,
		Offset:    (in.Page - 1) * in.Limit,
	})
	if err != nil {
		return ListOutput{}, err
	}

	totalPages := int(total / int64(in.Limit))
	if total%int64(in.Limit) != 0 {
		totalPages++
	}

	return ListOutput{
		Customers:  customers,
		Total:      total,
		Page:       in.Page,
		Limit:      in.Limit,
		TotalPages: totalPages,
	}, nil
}

func isValidSort(sort string) bool {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "", "created_at_desc", "created_at_asc", "points_desc", "points_asc":
		return true
	default:
		return false
	}
}

// pointsRangeForTier translates a tier name into the [min, max] points window.
// max = -1 means "no upper bound" (the top tier).
func pointsRangeForTier(name string) (min, max int, ok bool) {
	for i, t := range model.Tiers {
		if strings.EqualFold(t.Name, name) {
			if i+1 < len(model.Tiers) {
				return t.MinPoints, model.Tiers[i+1].MinPoints - 1, true
			}
			return t.MinPoints, -1, true
		}
	}
	return 0, 0, false
}
