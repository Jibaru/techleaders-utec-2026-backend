package purchase

import (
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"

	"mvc-coffee-loyalty/internal/controller/httpx"
	"mvc-coffee-loyalty/internal/model"
	customerview "mvc-coffee-loyalty/internal/view/customer"
	purchaseview "mvc-coffee-loyalty/internal/view/purchase"
)

var (
	errAlreadyRefunded    = errors.New("purchase already refunded")
	errPointsAlreadySpent = errors.New("points already spent")
)

func (c *Controller) Refund(w http.ResponseWriter, r *http.Request) {
	purchaseID, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid purchase id")
		return
	}

	var (
		updatedCustomer model.Customer
		updatedPurchase model.Purchase
		previousTier    string
		newTier         string
	)

	// Refund touches two rows that must move together: customer.points and
	// purchase.refunded_at. Wrap them in a transaction so a partial failure
	// cannot leave the points balance out of sync with the purchase log.
	err = c.db.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		var purchase model.Purchase
		if err := tx.First(&purchase, "id = ?", purchaseID).Error; err != nil {
			return err
		}
		if purchase.RefundedAt != nil {
			return errAlreadyRefunded
		}

		var customer model.Customer
		if err := tx.First(&customer, "id = ?", purchase.CustomerID).Error; err != nil {
			return err
		}

		if customer.Points < purchase.PointsEarned {
			return errPointsAlreadySpent
		}

		previousTier = model.TierForPoints(customer.Points).Name
		customer.Points -= purchase.PointsEarned
		now := time.Now().UTC()
		purchase.RefundedAt = &now

		if err := tx.Model(&customer).Update("points", customer.Points).Error; err != nil {
			return err
		}
		if err := tx.Model(&purchase).Update("refunded_at", purchase.RefundedAt).Error; err != nil {
			return err
		}

		newTier = model.TierForPoints(customer.Points).Name
		updatedCustomer = customer
		updatedPurchase = purchase
		return nil
	})

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		httpx.WriteError(w, http.StatusNotFound, "purchase not found")
		return
	case errors.Is(err, errAlreadyRefunded):
		httpx.WriteError(w, http.StatusConflict, "purchase already refunded")
		return
	case errors.Is(err, errPointsAlreadySpent):
		httpx.WriteError(w, http.StatusConflict, "cannot refund: customer has already spent those points")
		return
	case err != nil:
		httpx.WriteError(w, http.StatusInternalServerError, "could not refund purchase")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, purchaseview.RefundResponse{
		Purchase:       purchaseview.NewResponse(updatedPurchase),
		Customer:       customerview.NewResponse(updatedCustomer),
		PointsReverted: updatedPurchase.PointsEarned,
		PreviousTier:   previousTier,
		NewTier:        newTier,
		TierDemoted:    previousTier != newTier,
	})
}
