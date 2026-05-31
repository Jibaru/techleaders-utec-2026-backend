package view

import (
	"time"

	"github.com/google/uuid"

	"hexagonal-modular/internal/purchase/model"
	customerview "hexagonal-modular/internal/customer/view"
)

type Response struct {
	ID                uuid.UUID  `json:"id"`
	CustomerID        uuid.UUID  `json:"customer_id"`
	AmountCents       int64      `json:"amount_cents"`
	PointsEarned      int        `json:"points_earned"`
	ExternalPaymentID *string    `json:"external_payment_id,omitempty"`
	RefundedAt        *time.Time `json:"refunded_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type RecordedResponse struct {
	Purchase     Response              `json:"purchase"`
	Customer     customerview.Response `json:"customer"`
	PreviousTier string                `json:"previous_tier"`
	NewTier      string                `json:"new_tier"`
	TierUpgraded bool                  `json:"tier_upgraded"`
}

type RefundResponse struct {
	Purchase       Response              `json:"purchase"`
	Customer       customerview.Response `json:"customer"`
	PointsReverted int                   `json:"points_reverted"`
	PreviousTier   string                `json:"previous_tier"`
	NewTier        string                `json:"new_tier"`
	TierDemoted    bool                  `json:"tier_demoted"`
}

func NewResponse(p model.Purchase) Response {
	return Response{
		ID:                p.ID,
		CustomerID:        p.CustomerID,
		AmountCents:       p.AmountCents,
		PointsEarned:      p.PointsEarned,
		ExternalPaymentID: p.ExternalPaymentID,
		RefundedAt:        p.RefundedAt,
		CreatedAt:         p.CreatedAt,
	}
}
