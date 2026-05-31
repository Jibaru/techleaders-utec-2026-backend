package customer

import (
	"context"

	"github.com/google/uuid"

	"mvc-service-repo/internal/model"
)

func (s *Service) Get(ctx context.Context, id uuid.UUID) (model.Customer, error) {
	return s.customers.FindByID(ctx, id)
}
