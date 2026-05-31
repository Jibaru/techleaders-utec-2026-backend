package controller

import (
	"net/http"

	"hexagonal-modular-sidecar/internal/shared/httpx"
	purchaseservice "hexagonal-modular-sidecar/internal/purchase/service"
	purchaseview "hexagonal-modular-sidecar/internal/purchase/view"
	webhookview "hexagonal-modular-sidecar/internal/webhook/view"
)

func (c *Controller) PaymentConfirmed(w http.ResponseWriter, r *http.Request) {
	var req webhookview.PaymentConfirmed
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	customer, err := c.customers.GetByEmail(r.Context(), req.CustomerEmail)
	if err != nil {
		httpx.MapDomainError(w, err)
		return
	}

	// Same call the regular purchase endpoint makes. Idempotency by
	// external_payment_id is handled inside the service.
	result, err := c.purchases.Record(r.Context(), purchaseservice.RecordInput{
		CustomerID:        customer.ID,
		AmountCents:       req.AmountCents,
		ExternalPaymentID: &req.ExternalPaymentID,
		OccurredAt:        req.OccurredAt,
	})
	if err != nil {
		httpx.MapDomainError(w, err)
		return
	}

	status := "accepted"
	httpStatus := http.StatusAccepted
	if result.Duplicate {
		status = "already_processed"
		httpStatus = http.StatusOK
	}
	httpx.WriteJSON(w, httpStatus, webhookview.AcceptedResponse{
		Status:   status,
		Purchase: purchaseview.NewResponse(result.Purchase),
	})
}
