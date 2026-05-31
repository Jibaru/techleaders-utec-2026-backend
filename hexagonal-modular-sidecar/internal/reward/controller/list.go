package controller

import (
	"net/http"

	"hexagonal-modular-sidecar/internal/shared/httpx"
	rewardview "hexagonal-modular-sidecar/internal/reward/view"
)

func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	customerID, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	rewards, err := c.svc.List(r.Context(), customerID)
	if err != nil {
		httpx.MapDomainError(w, err)
		return
	}

	responses := make([]rewardview.Response, 0, len(rewards))
	for _, rw := range rewards {
		responses = append(responses, rewardview.NewResponse(rw))
	}
	httpx.WriteJSON(w, http.StatusOK, responses)
}
