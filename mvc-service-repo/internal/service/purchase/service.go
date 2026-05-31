// Package purchase holds the business logic for recording, listing, and
// refunding purchases. It depends only on repository ports and the
// Transactor port — no GORM types.
package purchase

import (
	"gorm.io/gorm"

	"mvc-service-repo/internal/mail"
	customerrepo "mvc-service-repo/internal/repository/customer"
	purchaserepo "mvc-service-repo/internal/repository/purchase"
)

type Service struct {
	db        *gorm.DB
	customers *customerrepo.Repository
	purchases *purchaserepo.Repository
	mailer    *mail.Sender
}

func New(db *gorm.DB, customers *customerrepo.Repository, purchases *purchaserepo.Repository, mailer *mail.Sender) *Service {
	return &Service{db: db, customers: customers, purchases: purchases, mailer: mailer}
}
