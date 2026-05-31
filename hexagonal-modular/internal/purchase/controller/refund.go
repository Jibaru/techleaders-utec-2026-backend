package controller

import (
	"net/http"

	"hexagonal-modular/internal/shared/httpx"
	customerview "hexagonal-modular/internal/customer/view"
	purchaseview "hexagonal-modular/internal/purchase/view"
)

func (c *Controller) Refund(w http.ResponseWriter, r *http.Request) {
	purchaseID, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid purchase id")
		return
	}

	result, err := c.svc.Refund(r.Context(), purchaseID)
	if err != nil {
		httpx.MapDomainError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, purchaseview.RefundResponse{
		Purchase:       purchaseview.NewResponse(result.Purchase),
		Customer:       customerview.NewResponse(result.Customer),
		PointsReverted: result.PointsReverted,
		PreviousTier:   result.PreviousTier.Name,
		NewTier:        result.NewTier.Name,
		TierDemoted:    result.PreviousTier.Name != result.NewTier.Name,
	})
}
