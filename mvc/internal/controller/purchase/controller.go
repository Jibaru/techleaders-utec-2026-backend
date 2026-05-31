// Package purchase holds the HTTP controller for purchases (create, list, refund).
package purchase

import "gorm.io/gorm"

type Controller struct {
	db *gorm.DB
}

func NewController(db *gorm.DB) *Controller {
	return &Controller{db: db}
}
