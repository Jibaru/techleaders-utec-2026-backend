// Package reward holds the HTTP controller for reward redemption and the reward catalog.
package reward

import (
	"gorm.io/gorm"

	"mvc-coffee-loyalty/internal/mail"
)

type Controller struct {
	db     *gorm.DB
	mailer *mail.Sender
}

func NewController(db *gorm.DB, mailer *mail.Sender) *Controller {
	return &Controller{db: db, mailer: mailer}
}
