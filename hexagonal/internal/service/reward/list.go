package reward

import (
	"context"

	"github.com/google/uuid"

	"hexagonal/internal/model"
)

func (s *Service) List(ctx context.Context, customerID uuid.UUID) ([]model.Reward, error) {
	exists, err := s.customers.ExistsByID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, model.ErrCustomerNotFound
	}
	return s.rewards.ListByCustomer(ctx, customerID)
}
