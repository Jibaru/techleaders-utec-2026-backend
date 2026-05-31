// Package purchase holds the business logic for recording, listing, and
// refunding purchases. It encodes the loyalty rules (point calculation, tier
// upgrades, refund reversal) and the idempotency check used by webhooks.
package purchase

import (
	"gorm.io/gorm"

	customerrepo "mvc-service-repo/internal/repository/customer"
	purchaserepo "mvc-service-repo/internal/repository/purchase"
)

type Service struct {
	db        *gorm.DB
	customers *customerrepo.Repository
	purchases *purchaserepo.Repository
}

func New(db *gorm.DB, customers *customerrepo.Repository, purchases *purchaserepo.Repository) *Service {
	return &Service{db: db, customers: customers, purchases: purchases}
}
