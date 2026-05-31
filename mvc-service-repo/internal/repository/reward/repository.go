// Package reward holds the persistence layer for the reward resource.
package reward

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mvc-service-repo/internal/model"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

func (r *Repository) Save(ctx context.Context, rw *model.Reward) error {
	return r.db.WithContext(ctx).Create(rw).Error
}

func (r *Repository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]model.Reward, error) {
	var rewards []model.Reward
	err := r.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Find(&rewards).Error
	return rewards, err
}

// Aggregate is the summed-up view of a customer's redeemed rewards.
type Aggregate struct {
	TotalSpent int64
	Count      int64
}

func (r *Repository) AggregateForCustomer(ctx context.Context, customerID uuid.UUID) (Aggregate, error) {
	var agg Aggregate
	err := r.db.WithContext(ctx).
		Model(&model.Reward{}).
		Where("customer_id = ?", customerID).
		Select(`COALESCE(SUM(points_spent), 0) AS total_spent,
		        COUNT(*)                       AS count`).
		Scan(&agg).Error
	return agg, err
}
