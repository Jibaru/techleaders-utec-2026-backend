package purchase

import (
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mvc-coffee-loyalty/internal/controller/httpx"
	"mvc-coffee-loyalty/internal/model"
	customerview "mvc-coffee-loyalty/internal/view/customer"
	purchaseview "mvc-coffee-loyalty/internal/view/purchase"
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
	if err := httpx.ValidateAmount(req.AmountCents); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
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

	previousTier := model.TierForPoints(customer.Points)
	dollars := float64(req.AmountCents) / 100.0
	pointsEarned := int(math.Floor(dollars * previousTier.Multiplier))

	purchase := model.Purchase{
		ID:           uuid.New(),
		CustomerID:   customer.ID,
		AmountCents:  req.AmountCents,
		PointsEarned: pointsEarned,
		CreatedAt:    time.Now().UTC(),
	}

	if err := c.db.WithContext(r.Context()).Create(&purchase).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not save purchase")
		return
	}

	customer.Points += pointsEarned
	if err := c.db.WithContext(r.Context()).
		Model(&model.Customer{}).
		Where("id = ?", customer.ID).
		Update("points", customer.Points).Error; err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not update customer points")
		return
	}

	newTier := model.TierForPoints(customer.Points)

	httpx.WriteJSON(w, http.StatusCreated, purchaseview.RecordedResponse{
		Purchase:     purchaseview.NewResponse(purchase),
		Customer:     customerview.NewResponse(customer),
		PreviousTier: previousTier.Name,
		NewTier:      newTier.Name,
		TierUpgraded: previousTier.Name != newTier.Name,
	})
}
