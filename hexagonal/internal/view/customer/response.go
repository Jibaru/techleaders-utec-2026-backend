package customer

import (
	"time"

	"github.com/google/uuid"

	"hexagonal/internal/model"
	"hexagonal/internal/view/shared"
	tierview "hexagonal/internal/view/tier"
)

type Response struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	Points    int            `json:"points"`
	Tier      tierview.Brief `json:"tier"`
	CreatedAt time.Time      `json:"created_at"`
}

type ListResponse struct {
	Data []Response      `json:"data"`
	Meta shared.ListMeta `json:"meta"`
}

type SummaryResponse struct {
	CustomerID        uuid.UUID       `json:"customer_id"`
	CurrentPoints     int             `json:"current_points"`
	CurrentTier       tierview.Brief  `json:"current_tier"`
	NextTier          *tierview.Brief `json:"next_tier,omitempty"`
	PointsToNextTier  int             `json:"points_to_next_tier"`
	TotalSpentCents   int64           `json:"total_spent_cents"`
	TotalPointsEarned int             `json:"total_points_earned"`
	TotalPointsSpent  int             `json:"total_points_spent"`
	PurchaseCount     int             `json:"purchase_count"`
	AvgPurchaseCents  int64           `json:"avg_purchase_cents"`
	RewardCount       int             `json:"reward_count"`
	LastPurchaseAt    *time.Time      `json:"last_purchase_at,omitempty"`
}

func NewResponse(c model.Customer) Response {
	return Response{
		ID:        c.ID,
		Name:      c.Name,
		Email:     c.Email,
		Points:    c.Points,
		Tier:      tierview.NewBrief(model.TierForPoints(c.Points)),
		CreatedAt: c.CreatedAt,
	}
}
