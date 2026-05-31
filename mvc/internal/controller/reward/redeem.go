package reward

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mvc-coffee-loyalty/internal/controller/httpx"
	"mvc-coffee-loyalty/internal/model"
	rewardview "mvc-coffee-loyalty/internal/view/reward"
	tierview "mvc-coffee-loyalty/internal/view/tier"
)

func (c *Controller) Redeem(w http.ResponseWriter, r *http.Request) {
	customerID, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	var req rewardview.RedeemRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rewardType := model.RewardType(req.Type)
	cost, ok := model.RewardCosts[rewardType]
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "unknown reward type")
		return
	}

	var customer model.Customer
	err = c.db.WithContext(r.Context()).First(&customer, "id = ?", customerID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "customer not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not load customer")
		return
	}

	if customer.Points < cost {
		httpx.WriteError(w, http.StatusConflict, "insufficient points to redeem reward")
		return
	}

	reward := model.Reward{
		ID:          uuid.New(),
		CustomerID:  customer.ID,
		Type:        rewardType,
		PointsSpent: cost,
		CreatedAt:   time.Now().UTC(),
	}

	if err := c.db.WithContext(r.Context()).Create(&reward).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not save reward")
		return
	}

	customer.Points -= cost
	if err := c.db.WithContext(r.Context()).
		Model(&model.Customer{}).
		Where("id = ?", customer.ID).
		Update("points", customer.Points).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not update customer points")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, rewardview.RedeemedResponse{
		Reward:          rewardview.NewResponse(reward),
		RemainingPoints: customer.Points,
		Tier:            tierview.NewBrief(model.TierForPoints(customer.Points)),
	})
}
