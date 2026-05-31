// Package purchase holds the business logic for recording, listing, and
// refunding purchases. It depends only on repository ports, the Transactor
// port, and the mail.Sender port — no GORM types, no SMTP.
package purchase

import (
	"hexagonal/internal/mail"
	"hexagonal/internal/repository"
	customerrepo "hexagonal/internal/repository/customer"
	purchaserepo "hexagonal/internal/repository/purchase"
)

type Service struct {
	customers customerrepo.Repository
	purchases purchaserepo.Repository
	tx        repository.Transactor
	mailer    mail.Sender
}

func New(
	customers customerrepo.Repository,
	purchases purchaserepo.Repository,
	tx repository.Transactor,
	mailer mail.Sender,
) *Service {
	return &Service{customers: customers, purchases: purchases, tx: tx, mailer: mailer}
}
