// Package customer holds the HTTP controller for the customer resource.
package customer

import customerservice "mvc-service-repo/internal/service/customer"

type Controller struct {
	svc *customerservice.Service
}

func NewController(svc *customerservice.Service) *Controller {
	return &Controller{svc: svc}
}
