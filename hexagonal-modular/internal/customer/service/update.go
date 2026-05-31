package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"hexagonal-modular/internal/customer/model"
	"hexagonal-modular/internal/shared/validate"
)

// UpdateInput is the business-level input for Update. Each field is optional;
// only non-nil fields are applied.
type UpdateInput struct {
	Name  *string
	Email *string
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpdateInput) (model.Customer, error) {
	updates := map[string]any{}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if err := validate.Name(name); err != nil {
			return model.Customer{}, err
		}
		updates["name"] = name
	}
	if in.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*in.Email))
		if err := validate.Email(email); err != nil {
			return model.Customer{}, err
		}
		updates["email"] = email
	}
	if len(updates) == 0 {
		return model.Customer{}, fmt.Errorf("%w: no fields to update", validate.ErrInvalidInput)
	}

	if _, err := s.customers.FindByID(ctx, id); err != nil {
		return model.Customer{}, err
	}
	return s.customers.Update(ctx, id, updates)
}
