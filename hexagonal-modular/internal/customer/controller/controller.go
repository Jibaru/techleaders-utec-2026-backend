// Package customer holds the HTTP controller for the customer resource.
package controller

import customerservice "hexagonal-modular/internal/customer/service"

type Controller struct {
	svc *customerservice.Service
}

func NewController(svc *customerservice.Service) *Controller {
	return &Controller{svc: svc}
}
