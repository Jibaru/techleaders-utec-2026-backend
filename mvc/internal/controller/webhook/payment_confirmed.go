package webhook

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mvc-coffee-loyalty/internal/controller/httpx"
	"mvc-coffee-loyalty/internal/model"
	purchaseview "mvc-coffee-loyalty/internal/view/purchase"
	webhookview "mvc-coffee-loyalty/internal/view/webhook"
)

func (c *Controller) PaymentConfirmed(w http.ResponseWriter, r *http.Request) {
	var req webhookview.PaymentConfirmed
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	externalID := strings.TrimSpace(req.ExternalPaymentID)
	email := strings.ToLower(strings.TrimSpace(req.CustomerEmail))
	if err := httpx.ValidateExternalID(externalID); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := httpx.ValidateEmail(email); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := httpx.ValidateAmount(req.AmountCents); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Idempotency: if we have already processed this external_payment_id,
	// return the existing purchase rather than creating a duplicate.
	var existing model.Purchase
	err := c.db.WithContext(r.Context()).
		Where("external_payment_id = ?", externalID).
		First(&existing).Error
	if err == nil {
		httpx.WriteJSON(w, http.StatusOK, webhookview.AcceptedResponse{
			Status:   "already_processed",
			Purchase: purchaseview.NewResponse(existing),
		})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		httpx.WriteError(w, http.StatusInternalServerError, "could not check idempotency")
		return
	}

	var customer model.Customer
	err = c.db.WithContext(r.Context()).First(&customer, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "customer not found for that email")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not load customer")
		return
	}

	// --- duplicated from purchase.Create ---
	previousTier := model.TierForPoints(customer.Points)
	dollars := float64(req.AmountCents) / 100.0
	pointsEarned := int(math.Floor(dollars * previousTier.Multiplier))

	createdAt := req.OccurredAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	purchase := model.Purchase{
		ID:                uuid.New(),
		CustomerID:        customer.ID,
		AmountCents:       req.AmountCents,
		PointsEarned:      pointsEarned,
		ExternalPaymentID: &externalID,
		CreatedAt:         createdAt,
	}

	err = c.db.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&purchase).Error; err != nil {
			return err
		}
		customer.Points += pointsEarned
		return tx.Model(&model.Customer{}).
			Where("id = ?", customer.ID).
			Update("points", customer.Points).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			// Race: a concurrent webhook processed the same external_payment_id.
			_ = c.db.WithContext(r.Context()).
				Where("external_payment_id = ?", externalID).
				First(&existing).Error
			httpx.WriteJSON(w, http.StatusOK, webhookview.AcceptedResponse{
				Status:   "already_processed",
				Purchase: purchaseview.NewResponse(existing),
			})
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "could not record purchase")
		return
	}

	// --- ALSO duplicated from purchase.Create: same receipt + tier upgrade emails ---
	newTier := model.TierForPoints(customer.Points)
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

	httpx.WriteJSON(w, http.StatusAccepted, webhookview.AcceptedResponse{
		Status:   "accepted",
		Purchase: purchaseview.NewResponse(purchase),
	})
}
