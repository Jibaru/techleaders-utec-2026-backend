// Package model holds the purchase entity and purchase-specific domain errors.
package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Purchase struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CustomerID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	AmountCents       int64      `gorm:"not null;check:amount_cents > 0"`
	PointsEarned      int        `gorm:"not null;check:points_earned >= 0"`
	ExternalPaymentID *string    `gorm:"type:varchar(128);uniqueIndex"`
	RefundedAt        *time.Time `gorm:"index"`
	CreatedAt         time.Time  `gorm:"not null;autoCreateTime"`
}

// Purchase-specific domain errors.
var (
	ErrPurchaseNotFound   = errors.New("purchase not found")
	ErrDuplicatePurchase  = errors.New("duplicate purchase for that external_payment_id")
	ErrAlreadyRefunded    = errors.New("purchase already refunded")
	ErrPointsAlreadySpent = errors.New("cannot refund: customer has already spent those points")
)
