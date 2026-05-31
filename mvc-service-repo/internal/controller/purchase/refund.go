package purchase

import (
	"net/http"

	"mvc-service-repo/internal/controller/httpx"
	customerview "mvc-service-repo/internal/view/customer"
	purchaseview "mvc-service-repo/internal/view/purchase"
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
