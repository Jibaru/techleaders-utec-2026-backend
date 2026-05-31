package model

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"not null"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Points    int       `gorm:"not null;default:0;check:points >= 0"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
}
