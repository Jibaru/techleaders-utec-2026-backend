package customer

import (
	"errors"
	"net/http"
	"strconv"

	"hexagonal/internal/controller/httpx"
	customerservice "hexagonal/internal/service/customer"
	customerview "hexagonal/internal/view/customer"
	"hexagonal/internal/view/shared"
)

func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, err := parsePositiveInt(q.Get("page"), 1)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid page")
		return
	}
	limit, err := parsePositiveInt(q.Get("limit"), 20)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid limit")
		return
	}

	in := customerservice.ListInput{
		Tier:  q.Get("tier"),
		Page:  page,
		Limit: limit,
		Sort:  q.Get("sort"),
	}
	if s := q.Get("min_points"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid min_points")
			return
		}
		in.MinPoints = &n
	}
	if s := q.Get("max_points"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid max_points")
			return
		}
		in.MaxPoints = &n
	}

	out, err := c.svc.List(r.Context(), in)
	if err != nil {
		httpx.MapDomainError(w, err)
		return
	}

	data := make([]customerview.Response, 0, len(out.Customers))
	for _, cu := range out.Customers {
		data = append(data, customerview.NewResponse(cu))
	}
	httpx.WriteJSON(w, http.StatusOK, customerview.ListResponse{
		Data: data,
		Meta: shared.ListMeta{Page: out.Page, Limit: out.Limit, Total: out.Total, TotalPages: out.TotalPages},
	})
}

func parsePositiveInt(s string, fallback int) (int, error) {
	if s == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 0, errors.New("not a positive integer")
	}
	return n, nil
}
