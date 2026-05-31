// Package customer holds the business logic for the customer resource.
// The service knows nothing about HTTP. It owns input cleaning (trim,
// lowercase) and format validation in addition to the business rules.
package customer

import (
	customerrepo "mvc-service-repo/internal/repository/customer"
	purchaserepo "mvc-service-repo/internal/repository/purchase"
	rewardrepo "mvc-service-repo/internal/repository/reward"
)

type Service struct {
	customers *customerrepo.Repository
	purchases *purchaserepo.Repository
	rewards   *rewardrepo.Repository
}

func New(customers *customerrepo.Repository, purchases *purchaserepo.Repository, rewards *rewardrepo.Repository) *Service {
	return &Service{customers: customers, purchases: purchases, rewards: rewards}
}
