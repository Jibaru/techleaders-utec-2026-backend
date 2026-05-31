package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"hexagonal-modular-sidecar/internal/customer/model"
	"hexagonal-modular-sidecar/internal/shared/validate"
)

func (s *Service) Create(ctx context.Context, rawName, rawEmail string) (model.Customer, error) {
	name := strings.TrimSpace(rawName)
	email := strings.ToLower(strings.TrimSpace(rawEmail))
	if err := validate.Name(name); err != nil {
		return model.Customer{}, err
	}
	if err := validate.Email(email); err != nil {
		return model.Customer{}, err
	}

	c := model.Customer{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		Points:    0,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.customers.Save(ctx, &c); err != nil {
		return model.Customer{}, err
	}
	return c, nil
}
