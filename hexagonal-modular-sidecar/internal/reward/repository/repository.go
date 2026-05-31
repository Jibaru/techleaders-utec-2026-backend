// Package reward holds the Repository port for the reward resource.
package repository

import (
	"context"

	"github.com/google/uuid"

	"hexagonal-modular-sidecar/internal/reward/model"
)

type Repository interface {
	Save(ctx context.Context, rw *model.Reward) error
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]model.Reward, error)
	AggregateForCustomer(ctx context.Context, customerID uuid.UUID) (Aggregate, error)
}

// Aggregate is the summed-up view of a customer's redeemed rewards.
type Aggregate struct {
	TotalSpent int64
	Count      int64
}
