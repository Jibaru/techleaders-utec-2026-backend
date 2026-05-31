// Package purchase holds the business logic for recording, listing, and
// refunding purchases. It depends only on repository ports and the
// Transactor port — no GORM types.
package purchase

import (
	"hexagonal/internal/repository"
	customerrepo "hexagonal/internal/repository/customer"
	purchaserepo "hexagonal/internal/repository/purchase"
)

type Service struct {
	customers customerrepo.Repository
	purchases purchaserepo.Repository
	tx        repository.Transactor
}

func New(customers customerrepo.Repository, purchases purchaserepo.Repository, tx repository.Transactor) *Service {
	return &Service{customers: customers, purchases: purchases, tx: tx}
}
