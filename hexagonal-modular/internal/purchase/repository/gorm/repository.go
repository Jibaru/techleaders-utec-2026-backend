// Package gorm holds the GORM adapter for the purchase Repository port.
package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	gormdb "gorm.io/gorm"

	"hexagonal-modular/internal/purchase/model"
	purchaserepo "hexagonal-modular/internal/purchase/repository"
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

func (r *Repository) Save(ctx context.Context, p *model.Purchase) error {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		if errors.Is(err, gormdb.ErrDuplicatedKey) {
			// Unique violation on external_payment_id — a concurrent
			// request beat us to it. The service treats this as idempotent.
			return model.ErrDuplicatePurchase
		}
		return err
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (model.Purchase, error) {
	var p model.Purchase
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return model.Purchase{}, model.ErrPurchaseNotFound
	}
	return p, err
}

func (r *Repository) FindByExternalPaymentID(ctx context.Context, externalID string) (model.Purchase, error) {
	var p model.Purchase
	err := r.db.WithContext(ctx).First(&p, "external_payment_id = ?", externalID).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return model.Purchase{}, model.ErrPurchaseNotFound
	}
	return p, err
}

func (r *Repository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]model.Purchase, error) {
	var purchases []model.Purchase
	err := r.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Find(&purchases).Error
	return purchases, err
}

func (r *Repository) MarkRefunded(ctx context.Context, id uuid.UUID, refundedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.Purchase{}).
		Where("id = ?", id).
		Update("refunded_at", refundedAt).Error
}

func (r *Repository) AggregateForCustomer(ctx context.Context, customerID uuid.UUID) (purchaserepo.Aggregate, error) {
	var agg purchaserepo.Aggregate
	err := r.db.WithContext(ctx).
		Model(&model.Purchase{}).
		Where("customer_id = ? AND refunded_at IS NULL", customerID).
		Select(`COALESCE(SUM(amount_cents), 0)  AS total_spent,
		        COALESCE(SUM(points_earned), 0) AS total_earned,
		        COUNT(*)                         AS count,
		        MAX(created_at)                  AS last_purchase_at`).
		Scan(&agg).Error
	return agg, err
}
