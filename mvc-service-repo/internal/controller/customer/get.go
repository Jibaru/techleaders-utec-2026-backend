package customer

import (
	"net/http"

	"mvc-service-repo/internal/controller/httpx"
	customerview "mvc-service-repo/internal/view/customer"
)

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	customer, err := c.svc.Get(r.Context(), id)
	if err != nil {
		httpx.MapDomainError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, customerview.NewResponse(customer))
}
