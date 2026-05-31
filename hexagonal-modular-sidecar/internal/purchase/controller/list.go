package controller

import (
	"net/http"

	"hexagonal-modular-sidecar/internal/shared/httpx"
	purchaseview "hexagonal-modular-sidecar/internal/purchase/view"
)

func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	customerID, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	purchases, err := c.svc.List(r.Context(), customerID)
	if err != nil {
		httpx.MapDomainError(w, err)
		return
	}

	responses := make([]purchaseview.Response, 0, len(purchases))
	for _, p := range purchases {
		responses = append(responses, purchaseview.NewResponse(p))
	}
	httpx.WriteJSON(w, http.StatusOK, responses)
}
