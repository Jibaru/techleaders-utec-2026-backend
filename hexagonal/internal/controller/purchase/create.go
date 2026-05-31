package purchase

import (
	"net/http"

	"hexagonal/internal/controller/httpx"
	purchaseservice "hexagonal/internal/service/purchase"
	customerview "hexagonal/internal/view/customer"
	purchaseview "hexagonal/internal/view/purchase"
)

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	customerID, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	var req purchaseview.CreateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.svc.Record(r.Context(), purchaseservice.RecordInput{
		CustomerID:  customerID,
		AmountCents: req.AmountCents,
	})
	if err != nil {
		httpx.MapDomainError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, purchaseview.RecordedResponse{
		Purchase:     purchaseview.NewResponse(result.Purchase),
		Customer:     customerview.NewResponse(result.Customer),
		PreviousTier: result.PreviousTier.Name,
		NewTier:      result.NewTier.Name,
		TierUpgraded: result.PreviousTier.Name != result.NewTier.Name,
	})
}
