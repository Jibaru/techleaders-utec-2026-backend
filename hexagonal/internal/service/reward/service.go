// Package reward holds the business logic for redeeming and listing rewards.
package reward

import (
	"hexagonal/internal/mail"
	"hexagonal/internal/repository"
	customerrepo "hexagonal/internal/repository/customer"
	rewardrepo "hexagonal/internal/repository/reward"
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
