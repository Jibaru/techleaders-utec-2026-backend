package controller

import (
	"net/http"

	"hexagonal-modular/internal/shared/httpx"
	customerview "hexagonal-modular/internal/customer/view"
)

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	var req customerview.CreateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	customer, err := c.svc.Create(r.Context(), req.Name, req.Email)
	if err != nil {
		httpx.MapDomainError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, customerview.NewResponse(customer))
}
