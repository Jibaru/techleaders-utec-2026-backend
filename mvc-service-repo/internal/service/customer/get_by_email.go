package customer

import (
	"context"
	"strings"

	"mvc-service-repo/internal/model"
	"mvc-service-repo/internal/validate"
)

func (s *Service) GetByEmail(ctx context.Context, rawEmail string) (model.Customer, error) {
	email := strings.ToLower(strings.TrimSpace(rawEmail))
	if err := validate.Email(email); err != nil {
		return model.Customer{}, err
	}
	return s.customers.FindByEmail(ctx, email)
}
