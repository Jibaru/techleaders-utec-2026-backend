// Package reward holds the HTTP controller for reward redemption, listing, and catalog.
package reward

import rewardservice "mvc-service-repo/internal/service/reward"

type Controller struct {
	svc *rewardservice.Service
}

func NewController(svc *rewardservice.Service) *Controller {
	return &Controller{svc: svc}
}
