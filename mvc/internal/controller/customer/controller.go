// Package customer holds the HTTP controller for the customer resource.
package customer

import "gorm.io/gorm"

type Controller struct {
	db *gorm.DB
}

func NewController(db *gorm.DB) *Controller {
	return &Controller{db: db}
}
