// Command seed wipes the loyalty tables and inserts a known set of fixtures
// so the API has interesting data to demo against. Run with `make seed`.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/gorm"

	"hexagonal/internal/config"
	"hexagonal/internal/db"
	"hexagonal/internal/model"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	gormDB, err := db.NewPostgresDB(ctx, logger, cfg.DB.DSN())
	if err != nil {
		logger.Error("connect db", "err", err)
		os.Exit(1)
	}

	// Ensure the schema exists so `make seed` can be the first command someone
	// runs against a fresh database.
	if err := gormDB.AutoMigrate(&model.Customer{}, &model.Purchase{}, &model.Reward{}); err != nil {
		logger.Error("auto migrate", "err", err)
		os.Exit(1)
	}

	if err := seed(ctx, gormDB, logger); err != nil {
		logger.Error("seed failed", "err", err)
		os.Exit(1)
	}
}

func seed(ctx context.Context, gormDB *gorm.DB, logger *slog.Logger) error {
	customers, purchases, rewards := buildFixtures()

	return gormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("TRUNCATE TABLE rewards, purchases, customers CASCADE").Error; err != nil {
			return fmt.Errorf("truncate: %w", err)
		}
		if err := tx.Create(&customers).Error; err != nil {
			return fmt.Errorf("insert customers: %w", err)
		}
		if err := tx.Create(&purchases).Error; err != nil {
			return fmt.Errorf("insert purchases: %w", err)
		}
		if err := tx.Create(&rewards).Error; err != nil {
			return fmt.Errorf("insert rewards: %w", err)
		}
		logger.Info("seed completed",
			"customers", len(customers),
			"purchases", len(purchases),
			"rewards", len(rewards),
		)
		return nil
	})
}
