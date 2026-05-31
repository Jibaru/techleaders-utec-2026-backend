package customer

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"mvc-coffee-loyalty/internal/controller/httpx"
	"mvc-coffee-loyalty/internal/model"
	customerview "mvc-coffee-loyalty/internal/view/customer"
	"mvc-coffee-loyalty/internal/view/shared"
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
	if limit > 100 {
		limit = 100
	}

	tx := c.db.WithContext(r.Context()).Model(&model.Customer{})

	if tierName := strings.ToLower(strings.TrimSpace(q.Get("tier"))); tierName != "" {
		min, max, ok := pointsRangeForTier(tierName)
		if !ok {
			httpx.WriteError(w, http.StatusBadRequest, "unknown tier filter")
			return
		}
		tx = tx.Where("points >= ?", min)
		if max >= 0 {
			tx = tx.Where("points <= ?", max)
		}
	}

	if minPointsStr := q.Get("min_points"); minPointsStr != "" {
		minPoints, err := strconv.Atoi(minPointsStr)
		if err != nil || minPoints < 0 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid min_points")
			return
		}
		tx = tx.Where("points >= ?", minPoints)
	}
	if maxPointsStr := q.Get("max_points"); maxPointsStr != "" {
		maxPoints, err := strconv.Atoi(maxPointsStr)
		if err != nil || maxPoints < 0 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid max_points")
			return
		}
		tx = tx.Where("points <= ?", maxPoints)
	}

	order, ok := orderClauseForSort(q.Get("sort"))
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid sort")
		return
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not count customers")
		return
	}

	var customers []model.Customer
	if err := tx.Order(order).Limit(limit).Offset((page - 1) * limit).Find(&customers).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not list customers")
		return
	}

	data := make([]customerview.Response, 0, len(customers))
	for _, cu := range customers {
		data = append(data, customerview.NewResponse(cu))
	}

	totalPages := int(total / int64(limit))
	if total%int64(limit) != 0 {
		totalPages++
	}

	httpx.WriteJSON(w, http.StatusOK, customerview.ListResponse{
		Data: data,
		Meta: shared.ListMeta{Page: page, Limit: limit, Total: total, TotalPages: totalPages},
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

func pointsRangeForTier(name string) (min, max int, ok bool) {
	for i, t := range model.Tiers {
		if strings.EqualFold(t.Name, name) {
			if i+1 < len(model.Tiers) {
				return t.MinPoints, model.Tiers[i+1].MinPoints - 1, true
			}
			return t.MinPoints, -1, true
		}
	}
	return 0, 0, false
}

func orderClauseForSort(sort string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "", "created_at_desc":
		return "created_at DESC", true
	case "created_at_asc":
		return "created_at ASC", true
	case "points_desc":
		return "points DESC, created_at DESC", true
	case "points_asc":
		return "points ASC, created_at DESC", true
	default:
		return "", false
	}
}
