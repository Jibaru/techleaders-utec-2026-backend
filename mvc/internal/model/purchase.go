package model

import (
	"time"

	"github.com/google/uuid"
)

type Purchase struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CustomerID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	AmountCents         int64      `gorm:"not null;check:amount_cents > 0"`
	PointsEarned        int        `gorm:"not null;check:points_earned >= 0"`
	ExternalPaymentID   *string    `gorm:"type:varchar(128);uniqueIndex"`
	RefundedAt          *time.Time `gorm:"index"`
	CreatedAt           time.Time  `gorm:"not null;autoCreateTime"`
}
