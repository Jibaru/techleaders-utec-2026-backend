package service

import (
	"context"

	"github.com/google/uuid"

	"hexagonal-modular/internal/customer/model"
)

func (s *Service) Get(ctx context.Context, id uuid.UUID) (model.Customer, error) {
	return s.customers.FindByID(ctx, id)
}
