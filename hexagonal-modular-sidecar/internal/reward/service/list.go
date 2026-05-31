package service

import (
	"context"

	"github.com/google/uuid"

	customermodel "hexagonal-modular-sidecar/internal/customer/model"
	"hexagonal-modular-sidecar/internal/reward/model"
)

func (s *Service) List(ctx context.Context, customerID uuid.UUID) ([]model.Reward, error) {
	exists, err := s.customers.ExistsByID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, customermodel.ErrCustomerNotFound
	}
	return s.rewards.ListByCustomer(ctx, customerID)
}
