// Package repository holds the Transactor port that lets services compose
// multi-row writes atomically without depending on any specific database
// driver. Adapters (e.g. gorm/) implement it.
package repository

import (
	"context"

	"hexagonal/internal/repository/customer"
	"hexagonal/internal/repository/purchase"
	"hexagonal/internal/repository/reward"
)

// TxRepos bundles tx-bound repository instances handed to the user of
// Transactor.RunTx. All three repos share the same underlying transaction —
// any write made through them succeeds or fails as one unit.
type TxRepos struct {
	Customers customer.Repository
	Purchases purchase.Repository
	Rewards   reward.Repository
}

// Transactor is the port for atomic multi-repository writes. The service
// layer depends only on this interface. Implementations (e.g. the GORM
// adapter) own the actual transaction lifecycle.
type Transactor interface {
	RunTx(ctx context.Context, fn func(repos TxRepos) error) error
}
