package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hexagonal/internal/config"
	"hexagonal/internal/controller/customer"
	"hexagonal/internal/controller/purchase"
	"hexagonal/internal/controller/reward"
	"hexagonal/internal/controller/tier"
	"hexagonal/internal/controller/webhook"
	"hexagonal/internal/db"
	"hexagonal/internal/mail"
	mailnoop "hexagonal/internal/mail/noop"
	mailsmtp "hexagonal/internal/mail/smtp"
	"hexagonal/internal/model"
	repogorm "hexagonal/internal/repository/gorm"
	customergorm "hexagonal/internal/repository/customer/gorm"
	purchasegorm "hexagonal/internal/repository/purchase/gorm"
	rewardgorm "hexagonal/internal/repository/reward/gorm"
	"hexagonal/internal/router"
	customersvc "hexagonal/internal/service/customer"
	purchasesvc "hexagonal/internal/service/purchase"
	rewardsvc "hexagonal/internal/service/reward"
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
	logger.Info("connected to postgres", "host", cfg.DB.Host, "db", cfg.DB.Name)

	if err := gormDB.AutoMigrate(&model.Customer{}, &model.Purchase{}, &model.Reward{}); err != nil {
		logger.Error("auto migrate", "err", err)
		os.Exit(1)
	}
	logger.Info("schema migrated")

	// Composition root: pick the adapters for every outbound port, then wire
	// services. This is the ONLY place that knows which adapters are real —
	// services depend purely on the ports defined in internal/repository and
	// internal/mail.
	customerRepo := customergorm.New(gormDB)
	purchaseRepo := purchasegorm.New(gormDB)
	rewardRepo := rewardgorm.New(gormDB)
	transactor := repogorm.NewTransactor(gormDB, customerRepo, purchaseRepo, rewardRepo)

	var mailer mail.Sender
	if cfg.Mail.Host == "" {
		mailer = mailnoop.New()
		logger.Info("smtp not configured; using noop mail sender")
	} else {
		mailer = mailsmtp.New(cfg.Mail.Host, cfg.Mail.Port, cfg.Mail.User, cfg.Mail.Password, cfg.Mail.From)
		logger.Info("smtp configured", "host", cfg.Mail.Host, "from", cfg.Mail.From)
	}

	customerService := customersvc.New(customerRepo, purchaseRepo, rewardRepo)
	purchaseService := purchasesvc.New(customerRepo, purchaseRepo, transactor, mailer)
	rewardService := rewardsvc.New(customerRepo, rewardRepo, transactor, mailer)

	handler := router.New(router.Controllers{
		Customer: customer.NewController(customerService),
		Purchase: purchase.NewController(purchaseService),
		Reward:   reward.NewController(rewardService),
		Tier:     tier.NewController(),
		Webhook:  webhook.NewController(customerService, purchaseService),
	})
	handler = requestLogger(logger)(recoverer(logger)(handler))

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		logger.Error("http server failed", "err", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
	}
	logger.Info("bye")
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info("http",
				"method", r.Method,
				"path", r.URL.Path,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

func recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered", "panic", rec, "path", r.URL.Path)
					http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
