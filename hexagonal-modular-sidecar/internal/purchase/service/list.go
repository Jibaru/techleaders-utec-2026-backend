package service

import (
	"context"

	"github.com/google/uuid"

	customermodel "hexagonal-modular-sidecar/internal/customer/model"
	"hexagonal-modular-sidecar/internal/purchase/model"
)

func (s *Service) List(ctx context.Context, customerID uuid.UUID) ([]model.Purchase, error) {
	exists, err := s.customers.ExistsByID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, customermodel.ErrCustomerNotFound
	}
	return s.purchases.ListByCustomer(ctx, customerID)
}
