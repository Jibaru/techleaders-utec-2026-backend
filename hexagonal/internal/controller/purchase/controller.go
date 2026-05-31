// Package purchase holds the HTTP controller for purchases (create, list, refund).
package purchase

import purchaseservice "hexagonal/internal/service/purchase"

type Controller struct {
	svc *purchaseservice.Service
}

func NewController(svc *purchaseservice.Service) *Controller {
	return &Controller{svc: svc}
}
