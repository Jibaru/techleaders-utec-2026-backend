// Package model holds the reward entity, the reward catalog, and
// reward-specific domain errors.
package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type RewardType string

const (
	RewardFreeDrink  RewardType = "free_drink"
	RewardFreePastry RewardType = "free_pastry"
)

// RewardCosts is the static reward catalog. Lives in the reward module
// because no other module needs to know reward economics.
var RewardCosts = map[RewardType]int{
	RewardFreeDrink:  100,
	RewardFreePastry: 75,
}

type Reward struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CustomerID  uuid.UUID  `gorm:"type:uuid;not null;index"`
	Type        RewardType `gorm:"type:varchar(32);not null"`
	PointsSpent int        `gorm:"not null;check:points_spent > 0"`
	CreatedAt   time.Time  `gorm:"not null;autoCreateTime"`
}

// Reward-specific domain errors.
var (
	ErrInsufficientPoints = errors.New("insufficient points to redeem reward")
	ErrUnknownReward      = errors.New("unknown reward type")
)
