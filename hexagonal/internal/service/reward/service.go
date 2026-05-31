// Package reward holds the business logic for redeeming and listing rewards.
package reward

import (
	"hexagonal/internal/repository"
	customerrepo "hexagonal/internal/repository/customer"
	rewardrepo "hexagonal/internal/repository/reward"
)

type Service struct {
	customers customerrepo.Repository
	rewards   rewardrepo.Repository
	tx        repository.Transactor
}

func New(customers customerrepo.Repository, rewards rewardrepo.Repository, tx repository.Transactor) *Service {
	return &Service{customers: customers, rewards: rewards, tx: tx}
}
