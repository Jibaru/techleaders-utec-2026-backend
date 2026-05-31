package model

import (
	"time"

	"github.com/google/uuid"
)

type RewardType string

const (
	RewardFreeDrink  RewardType = "free_drink"
	RewardFreePastry RewardType = "free_pastry"
)

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
