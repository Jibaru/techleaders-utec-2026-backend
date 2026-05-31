package customer

import (
	"context"
	"strings"

	"hexagonal/internal/model"
	"hexagonal/internal/validate"
)

func (s *Service) GetByEmail(ctx context.Context, rawEmail string) (model.Customer, error) {
	email := strings.ToLower(strings.TrimSpace(rawEmail))
	if err := validate.Email(email); err != nil {
		return model.Customer{}, err
	}
	return s.customers.FindByEmail(ctx, email)
}
