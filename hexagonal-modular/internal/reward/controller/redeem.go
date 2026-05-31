package controller

import (
	"net/http"

	"hexagonal-modular/internal/shared/httpx"
	"hexagonal-modular/internal/reward/model"
	rewardview "hexagonal-modular/internal/reward/view"
	tierview "hexagonal-modular/internal/tier/view"
)

func (c *Controller) Redeem(w http.ResponseWriter, r *http.Request) {
	customerID, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	var req rewardview.RedeemRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.svc.Redeem(r.Context(), customerID, model.RewardType(req.Type))
	if err != nil {
		httpx.MapDomainError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, rewardview.RedeemedResponse{
		Reward:          rewardview.NewResponse(result.Reward),
		RemainingPoints: result.RemainingPoints,
		Tier:            tierview.NewBrief(result.Tier),
	})
}
