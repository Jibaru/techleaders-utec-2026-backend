// Package purchase holds the business logic for recording, listing, and
// refunding purchases. It depends only on repository ports, the Transactor
// port, and the mail.Sender port — no GORM types, no SMTP.
package service

import (
	"hexagonal-modular-sidecar/internal/shared/mail"
	"hexagonal-modular-sidecar/internal/shared/repository"
	customerrepo "hexagonal-modular-sidecar/internal/customer/repository"
	purchaserepo "hexagonal-modular-sidecar/internal/purchase/repository"
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
