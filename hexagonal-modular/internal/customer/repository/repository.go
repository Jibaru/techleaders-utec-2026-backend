// Package customer holds the Repository port (interface) for the customer
// resource. Implementations live in sub-packages (e.g. gorm/). The service
// layer depends only on this interface.
package repository

import (
	"context"

	"github.com/google/uuid"

	"hexagonal-modular/internal/customer/model"
)

// Repository is the port for customer persistence. Any adapter (GORM,
// in-memory, etc.) implements it.
type Repository interface {
	Save(ctx context.Context, c *model.Customer) error
	FindByID(ctx context.Context, id uuid.UUID) (model.Customer, error)
	FindByEmail(ctx context.Context, email string) (model.Customer, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	UpdatePoints(ctx context.Context, id uuid.UUID, points int) error
	Update(ctx context.Context, id uuid.UUID, updates map[string]any) (model.Customer, error)
	List(ctx context.Context, f ListFilter) ([]model.Customer, int64, error)
}

// ListFilter is the persistence-layer filter spec for List. Lives on the
// port so callers and adapters agree on the shape without sharing impl code.
type ListFilter struct {
	MinPoints *int
	MaxPoints *int
	Sort      string // "" / "created_at_desc" (default), "created_at_asc", "points_desc", "points_asc"
	Limit     int
	Offset    int
}
