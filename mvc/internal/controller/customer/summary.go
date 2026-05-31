package customer

import (
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"

	"mvc-coffee-loyalty/internal/controller/httpx"
	"mvc-coffee-loyalty/internal/model"
	customerview "mvc-coffee-loyalty/internal/view/customer"
	tierview "mvc-coffee-loyalty/internal/view/tier"
)

func (c *Controller) Summary(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	var customer model.Customer
	err = c.db.WithContext(r.Context()).First(&customer, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "customer not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not load customer")
		return
	}

	var purchaseAgg struct {
		TotalSpent     int64
		TotalEarned    int64
		PurchaseCount  int64
		LastPurchaseAt *time.Time
	}
	err = c.db.WithContext(r.Context()).
		Model(&model.Purchase{}).
		Where("customer_id = ? AND refunded_at IS NULL", id).
		Select(`COALESCE(SUM(amount_cents), 0)  AS total_spent,
		        COALESCE(SUM(points_earned), 0) AS total_earned,
		        COUNT(*)                         AS purchase_count,
		        MAX(created_at)                  AS last_purchase_at`).
		Scan(&purchaseAgg).Error
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not aggregate purchases")
		return
	}

	var rewardAgg struct {
		TotalSpent  int64
		RewardCount int64
	}
	err = c.db.WithContext(r.Context()).
		Model(&model.Reward{}).
		Where("customer_id = ?", id).
		Select(`COALESCE(SUM(points_spent), 0) AS total_spent,
		        COUNT(*)                       AS reward_count`).
		Scan(&rewardAgg).Error
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not aggregate rewards")
		return
	}

	currentTier := model.TierForPoints(customer.Points)
	nextTier := model.NextTierForPoints(customer.Points)
	pointsToNext := 0
	var nextBrief *tierview.Brief
	if nextTier != nil {
		pointsToNext = nextTier.MinPoints - customer.Points
		if pointsToNext < 0 {
			pointsToNext = 0
		}
		brief := tierview.NewBrief(*nextTier)
		nextBrief = &brief
	}

	var avg int64
	if purchaseAgg.PurchaseCount > 0 {
		avg = purchaseAgg.TotalSpent / purchaseAgg.PurchaseCount
	}

	httpx.WriteJSON(w, http.StatusOK, customerview.SummaryResponse{
		CustomerID:        customer.ID,
		CurrentPoints:     customer.Points,
		CurrentTier:       tierview.NewBrief(currentTier),
		NextTier:          nextBrief,
		PointsToNextTier:  pointsToNext,
		TotalSpentCents:   purchaseAgg.TotalSpent,
		TotalPointsEarned: int(purchaseAgg.TotalEarned),
		TotalPointsSpent:  int(rewardAgg.TotalSpent),
		PurchaseCount:     int(purchaseAgg.PurchaseCount),
		AvgPurchaseCents:  avg,
		RewardCount:       int(rewardAgg.RewardCount),
		LastPurchaseAt:    purchaseAgg.LastPurchaseAt,
	})
}
