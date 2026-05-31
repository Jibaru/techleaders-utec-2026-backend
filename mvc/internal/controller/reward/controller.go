// Package reward holds the HTTP controller for reward redemption and the reward catalog.
package reward

import "gorm.io/gorm"

type Controller struct {
	db *gorm.DB
}

func NewController(db *gorm.DB) *Controller {
	return &Controller{db: db}
}
