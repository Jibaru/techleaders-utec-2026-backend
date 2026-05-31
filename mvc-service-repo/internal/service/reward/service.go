// Package reward holds the business logic for redeeming and listing rewards.
package reward

import (
	"gorm.io/gorm"

	"mvc-service-repo/internal/mail"
	customerrepo "mvc-service-repo/internal/repository/customer"
	rewardrepo "mvc-service-repo/internal/repository/reward"
)

type Service struct {
	db        *gorm.DB
	customers *customerrepo.Repository
	rewards   *rewardrepo.Repository
	mailer    *mail.Sender
}

func New(db *gorm.DB, customers *customerrepo.Repository, rewards *rewardrepo.Repository, mailer *mail.Sender) *Service {
	return &Service{db: db, customers: customers, rewards: rewards, mailer: mailer}
}
