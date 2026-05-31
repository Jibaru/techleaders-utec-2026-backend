// Package purchase holds the HTTP controller for purchases (create, list, refund).
package controller

import purchaseservice "hexagonal-modular-sidecar/internal/purchase/service"

type Controller struct {
	svc *purchaseservice.Service
}

func NewController(svc *purchaseservice.Service) *Controller {
	return &Controller{svc: svc}
}
