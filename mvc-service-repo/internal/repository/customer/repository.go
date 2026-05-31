// Package customer holds the persistence layer for the customer resource.
// It owns all GORM queries against the customers table and translates
// infrastructure errors (pgx / gorm) into domain errors from the model package.
package customer

import (
	"context"
	"errors"

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

// WithTx returns a Repository bound to the given transaction. The service
// layer uses this to compose multi-row writes inside a gorm.Transaction.
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

func (r *Repository) Save(ctx context.Context, c *model.Customer) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return model.ErrEmailAlreadyExists
		}
		return err
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (model.Customer, error) {
	var c model.Customer
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Customer{}, model.ErrCustomerNotFound
	}
	return c, err
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (model.Customer, error) {
	var c model.Customer
	err := r.db.WithContext(ctx).First(&c, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Customer{}, model.ErrCustomerNotFound
	}
	return c, err
}

func (r *Repository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Customer{}).
		Where("id = ?", id).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) UpdatePoints(ctx context.Context, id uuid.UUID, points int) error {
	return r.db.WithContext(ctx).
		Model(&model.Customer{}).
		Where("id = ?", id).
		Update("points", points).Error
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (model.Customer, error) {
	tx := r.db.WithContext(ctx).Model(&model.Customer{}).Where("id = ?", id)
	if err := tx.Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return model.Customer{}, model.ErrEmailAlreadyExists
		}
		return model.Customer{}, err
	}
	return r.FindByID(ctx, id)
}

// ListFilter is the persistence-layer filter spec for List.
type ListFilter struct {
	MinPoints *int
	MaxPoints *int
	Sort      string // "created_at_desc" (default), "created_at_asc", "points_desc", "points_asc"
	Limit     int
	Offset    int
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]model.Customer, int64, error) {
	tx := r.db.WithContext(ctx).Model(&model.Customer{})
	if f.MinPoints != nil {
		tx = tx.Where("points >= ?", *f.MinPoints)
	}
	if f.MaxPoints != nil {
		tx = tx.Where("points <= ?", *f.MaxPoints)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var customers []model.Customer
	if err := tx.Order(orderClause(f.Sort)).
		Limit(f.Limit).
		Offset(f.Offset).
		Find(&customers).Error; err != nil {
		return nil, 0, err
	}
	return customers, total, nil
}

func orderClause(sort string) string {
	switch sort {
	case "created_at_asc":
		return "created_at ASC"
	case "points_desc":
		return "points DESC, created_at DESC"
	case "points_asc":
		return "points ASC, created_at DESC"
	default:
		return "created_at DESC"
	}
}
