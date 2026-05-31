// Package purchase holds the Repository port for the purchase resource.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"hexagonal-modular-sidecar/internal/purchase/model"
)

type Repository interface {
	Save(ctx context.Context, p *model.Purchase) error
	FindByID(ctx context.Context, id uuid.UUID) (model.Purchase, error)
	FindByExternalPaymentID(ctx context.Context, externalID string) (model.Purchase, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]model.Purchase, error)
	MarkRefunded(ctx context.Context, id uuid.UUID, refundedAt time.Time) error
	AggregateForCustomer(ctx context.Context, customerID uuid.UUID) (Aggregate, error)
}

// Aggregate is the summed-up view of a customer's non-refunded purchases.
type Aggregate struct {
	TotalSpent     int64
	TotalEarned    int64
	Count          int64
	LastPurchaseAt *time.Time
}
