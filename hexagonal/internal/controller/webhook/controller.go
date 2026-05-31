// Package webhook holds the HTTP controllers for inbound webhooks from
// external systems (e.g. a payment processor). With the service layer
// extracted, the webhook is now a thin adapter: it resolves the customer
// by email and forwards to the same purchase.Service.Record call that the
// regular API endpoint uses — no duplicated business logic.
package webhook

import (
	customerservice "hexagonal/internal/service/customer"
	purchaseservice "hexagonal/internal/service/purchase"
)

type Controller struct {
	customers *customerservice.Service
	purchases *purchaseservice.Service
}

func NewController(customers *customerservice.Service, purchases *purchaseservice.Service) *Controller {
	return &Controller{customers: customers, purchases: purchases}
}
