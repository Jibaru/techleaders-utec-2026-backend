package customer

import (
	"net/http"

	"hexagonal/internal/controller/httpx"
	customerview "hexagonal/internal/view/customer"
	tierview "hexagonal/internal/view/tier"
)

func (c *Controller) Summary(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	summary, err := c.svc.GetSummary(r.Context(), id)
	if err != nil {
		httpx.MapDomainError(w, err)
		return
	}

	var nextBrief *tierview.Brief
	if summary.NextTier != nil {
		b := tierview.NewBrief(*summary.NextTier)
		nextBrief = &b
	}

	httpx.WriteJSON(w, http.StatusOK, customerview.SummaryResponse{
		CustomerID:        summary.Customer.ID,
		CurrentPoints:     summary.Customer.Points,
		CurrentTier:       tierview.NewBrief(summary.CurrentTier),
		NextTier:          nextBrief,
		PointsToNextTier:  summary.PointsToNextTier,
		TotalSpentCents:   summary.TotalSpentCents,
		TotalPointsEarned: summary.TotalPointsEarned,
		TotalPointsSpent:  summary.TotalPointsSpent,
		PurchaseCount:     summary.PurchaseCount,
		AvgPurchaseCents:  summary.AvgPurchaseCents,
		RewardCount:       summary.RewardCount,
		LastPurchaseAt:    summary.LastPurchaseAt,
	})
}
