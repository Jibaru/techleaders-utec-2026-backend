// Package reward holds the HTTP controller for reward redemption, listing, and catalog.
package controller

import rewardservice "hexagonal-modular-sidecar/internal/reward/service"

type Controller struct {
	svc *rewardservice.Service
}

func NewController(svc *rewardservice.Service) *Controller {
	return &Controller{svc: svc}
}
