package purchase

import (
	"errors"
	"fmt"
	"log/slog"
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

	// Send the purchase receipt. Log-and-continue on failure: the DB write
	// already committed, so a flaky SMTP run shouldn't fail the request.
	receiptSubject := "Thanks for your purchase"
	receiptBody := fmt.Sprintf(
		"Hi %s,\n\nThanks for your $%.2f purchase. You earned %d points!\nYour balance is now %d points (%s).\n\n— Tu Café",
		customer.Name, dollars, pointsEarned, customer.Points, newTier.Name,
	)
	if err := c.mailer.Send(customer.Email, receiptSubject, receiptBody); err != nil {
		slog.Error("send purchase receipt", "err", err, "customer_id", customer.ID)
	}

	if previousTier.Name != newTier.Name {
		tierSubject := fmt.Sprintf("Welcome to %s!", newTier.Name)
		tierBody := fmt.Sprintf(
			"Hi %s,\n\nCongratulations — you've reached our %s tier!\nYou now earn points at %.2fx on every purchase.\n\n— Tu Café",
			customer.Name, newTier.Name, newTier.Multiplier,
		)
		if err := c.mailer.Send(customer.Email, tierSubject, tierBody); err != nil {
			slog.Error("send tier upgrade", "err", err, "customer_id", customer.ID)
		}
	}

	httpx.WriteJSON(w, http.StatusCreated, purchaseview.RecordedResponse{
		Purchase:     purchaseview.NewResponse(purchase),
		Customer:     customerview.NewResponse(customer),
		PreviousTier: previousTier.Name,
		NewTier:      newTier.Name,
		TierUpgraded: previousTier.Name != newTier.Name,
	})
}
