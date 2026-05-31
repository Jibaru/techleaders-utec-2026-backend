// Package purchase holds the HTTP controller for purchases (create, list, refund).
package purchase

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
