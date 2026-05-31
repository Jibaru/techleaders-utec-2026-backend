package purchase

import (
	"net/http"

	"mvc-coffee-loyalty/internal/controller/httpx"
	"mvc-coffee-loyalty/internal/model"
	purchaseview "mvc-coffee-loyalty/internal/view/purchase"
)

func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	customerID, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	var count int64
	if err := c.db.WithContext(r.Context()).
		Model(&model.Customer{}).
		Where("id = ?", customerID).
		Count(&count).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not load customer")
		return
	}
	if count == 0 {
		httpx.WriteError(w, http.StatusNotFound, "customer not found")
		return
	}

	var purchases []model.Purchase
	if err := c.db.WithContext(r.Context()).
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Find(&purchases).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not list purchases")
		return
	}

	responses := make([]purchaseview.Response, 0, len(purchases))
	for _, p := range purchases {
		responses = append(responses, purchaseview.NewResponse(p))
	}
	httpx.WriteJSON(w, http.StatusOK, responses)
}
