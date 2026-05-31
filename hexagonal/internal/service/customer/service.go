// Package customer holds the business logic for the customer resource.
// The service depends only on repository ports (interfaces) — never on
// concrete adapters or *gorm.DB. This is what makes it unit-testable.
package customer

import (
	customerrepo "hexagonal/internal/repository/customer"
	purchaserepo "hexagonal/internal/repository/purchase"
	rewardrepo "hexagonal/internal/repository/reward"
)

type Service struct {
	customers customerrepo.Repository
	purchases purchaserepo.Repository
	rewards   rewardrepo.Repository
}

func New(customers customerrepo.Repository, purchases purchaserepo.Repository, rewards rewardrepo.Repository) *Service {
	return &Service{customers: customers, purchases: purchases, rewards: rewards}
}
