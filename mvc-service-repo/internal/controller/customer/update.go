package customer

import (
	"net/http"

	"mvc-service-repo/internal/controller/httpx"
	customerservice "mvc-service-repo/internal/service/customer"
	customerview "mvc-service-repo/internal/view/customer"
)

func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	var req customerview.UpdateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	customer, err := c.svc.Update(r.Context(), id, customerservice.UpdateInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		httpx.MapDomainError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, customerview.NewResponse(customer))
}
