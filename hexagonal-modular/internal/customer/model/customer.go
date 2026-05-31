// Package model holds the customer entity and customer-specific domain errors.
package model

import (
	"errors"
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

// Customer-specific domain errors. Returned by the repository (after
// translating infra errors) and propagated by the service. Controllers map
// them to HTTP status codes via shared/httpx.MapDomainError.
var (
	ErrCustomerNotFound   = errors.New("customer not found")
	ErrEmailAlreadyExists = errors.New("email already registered")
)
