// Package purchase holds the persistence layer for the purchase resource.
package purchase

import (
	"context"
	"errors"
	"time"

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

func (r *Repository) Save(ctx context.Context, p *model.Purchase) error {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Purchase{}, model.ErrPurchaseNotFound
	}
	return p, err
}

// FindByExternalPaymentID returns the purchase tied to a payment processor id,
// or model.ErrPurchaseNotFound if no row exists.
func (r *Repository) FindByExternalPaymentID(ctx context.Context, externalID string) (model.Purchase, error) {
	var p model.Purchase
	err := r.db.WithContext(ctx).First(&p, "external_payment_id = ?", externalID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
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

// Aggregate is the summed-up view of a customer's non-refunded purchases.
type Aggregate struct {
	TotalSpent     int64
	TotalEarned    int64
	Count          int64
	LastPurchaseAt *time.Time
}

func (r *Repository) AggregateForCustomer(ctx context.Context, customerID uuid.UUID) (Aggregate, error) {
	var agg Aggregate
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
