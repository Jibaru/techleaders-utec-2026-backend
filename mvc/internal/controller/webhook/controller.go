// Package webhook holds the HTTP controllers for inbound webhooks from
// external systems (e.g. a payment processor). It is a second entry point that
// must apply the same purchase rules as the customer-facing API — and the
// resulting duplication is the punchline for extracting a service layer.
package webhook

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
