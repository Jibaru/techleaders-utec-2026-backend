package reward

import (
	"net/http"

	"mvc-coffee-loyalty/internal/controller/httpx"
	"mvc-coffee-loyalty/internal/model"
	rewardview "mvc-coffee-loyalty/internal/view/reward"
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

	var rewards []model.Reward
	if err := c.db.WithContext(r.Context()).
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Find(&rewards).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not list rewards")
		return
	}

	responses := make([]rewardview.Response, 0, len(rewards))
	for _, rw := range rewards {
		responses = append(responses, rewardview.NewResponse(rw))
	}
	httpx.WriteJSON(w, http.StatusOK, responses)
}
