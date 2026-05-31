// Package gorm holds the GORM adapter for the reward Repository port.
package gorm

import (
	"context"

	"github.com/google/uuid"
	gormdb "gorm.io/gorm"

	"hexagonal-modular/internal/reward/model"
	rewardrepo "hexagonal-modular/internal/reward/repository"
)

type Repository struct {
	db *gormdb.DB
}

func New(db *gormdb.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithTx(tx *gormdb.DB) *Repository {
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

func (r *Repository) AggregateForCustomer(ctx context.Context, customerID uuid.UUID) (rewardrepo.Aggregate, error) {
	var agg rewardrepo.Aggregate
	err := r.db.WithContext(ctx).
		Model(&model.Reward{}).
		Where("customer_id = ?", customerID).
		Select(`COALESCE(SUM(points_spent), 0) AS total_spent,
		        COUNT(*)                       AS count`).
		Scan(&agg).Error
	return agg, err
}
