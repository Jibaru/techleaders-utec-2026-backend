package purchase

import (
	"context"

	"github.com/google/uuid"

	"mvc-service-repo/internal/model"
)

func (s *Service) List(ctx context.Context, customerID uuid.UUID) ([]model.Purchase, error) {
	exists, err := s.customers.ExistsByID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, model.ErrCustomerNotFound
	}
	return s.purchases.ListByCustomer(ctx, customerID)
}
