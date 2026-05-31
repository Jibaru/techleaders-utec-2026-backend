// Package gorm holds the GORM adapter for the Transactor port.
package gorm

import (
	"context"

	gormdb "gorm.io/gorm"

	"hexagonal-modular-sidecar/internal/shared/repository"
	customergorm "hexagonal-modular-sidecar/internal/customer/repository/gorm"
	purchasegorm "hexagonal-modular-sidecar/internal/purchase/repository/gorm"
	rewardgorm "hexagonal-modular-sidecar/internal/reward/repository/gorm"
)

// Transactor implements repository.Transactor using gorm.DB.Transaction.
// Each call to RunTx opens a transaction, binds tx-aware copies of every
// repository, and passes them to the user-supplied function as TxRepos.
type Transactor struct {
	db        *gormdb.DB
	customers *customergorm.Repository
	purchases *purchasegorm.Repository
	rewards   *rewardgorm.Repository
}

func NewTransactor(
	db *gormdb.DB,
	customers *customergorm.Repository,
	purchases *purchasegorm.Repository,
	rewards *rewardgorm.Repository,
) *Transactor {
	return &Transactor{db: db, customers: customers, purchases: purchases, rewards: rewards}
}

func (t *Transactor) RunTx(ctx context.Context, fn func(repos repository.TxRepos) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gormdb.DB) error {
		return fn(repository.TxRepos{
			Customers: t.customers.WithTx(tx),
			Purchases: t.purchases.WithTx(tx),
			Rewards:   t.rewards.WithTx(tx),
		})
	})
}
