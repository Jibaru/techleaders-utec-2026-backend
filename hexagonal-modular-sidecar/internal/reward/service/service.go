// Package reward holds the business logic for redeeming and listing rewards.
package service

import (
	"hexagonal-modular-sidecar/internal/shared/mail"
	"hexagonal-modular-sidecar/internal/shared/repository"
	customerrepo "hexagonal-modular-sidecar/internal/customer/repository"
	rewardrepo "hexagonal-modular-sidecar/internal/reward/repository"
)

type Service struct {
	customers customerrepo.Repository
	rewards   rewardrepo.Repository
	tx        repository.Transactor
	mailer    mail.Sender
}

func New(
	customers customerrepo.Repository,
	rewards rewardrepo.Repository,
	tx repository.Transactor,
	mailer mail.Sender,
) *Service {
	return &Service{customers: customers, rewards: rewards, tx: tx, mailer: mailer}
}
