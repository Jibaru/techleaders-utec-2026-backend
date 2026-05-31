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

	customerctl "hexagonal-modular/internal/customer/controller"
	customermodel "hexagonal-modular/internal/customer/model"
	customergorm "hexagonal-modular/internal/customer/repository/gorm"
	customerrouter "hexagonal-modular/internal/customer/router"
	customersvc "hexagonal-modular/internal/customer/service"
	purchasectl "hexagonal-modular/internal/purchase/controller"
	purchasemodel "hexagonal-modular/internal/purchase/model"
	purchasegorm "hexagonal-modular/internal/purchase/repository/gorm"
	purchaserouter "hexagonal-modular/internal/purchase/router"
	purchasesvc "hexagonal-modular/internal/purchase/service"
	rewardctl "hexagonal-modular/internal/reward/controller"
	rewardmodel "hexagonal-modular/internal/reward/model"
	rewardgorm "hexagonal-modular/internal/reward/repository/gorm"
	rewardrouter "hexagonal-modular/internal/reward/router"
	rewardsvc "hexagonal-modular/internal/reward/service"
	"hexagonal-modular/internal/shared/config"
	"hexagonal-modular/internal/shared/db"
	"hexagonal-modular/internal/shared/mail"
	mailnoop "hexagonal-modular/internal/shared/mail/noop"
	mailsmtp "hexagonal-modular/internal/shared/mail/smtp"
	repogorm "hexagonal-modular/internal/shared/repository/gorm"
	tierctl "hexagonal-modular/internal/tier/controller"
	tierrouter "hexagonal-modular/internal/tier/router"
	webhookctl "hexagonal-modular/internal/webhook/controller"
	webhookrouter "hexagonal-modular/internal/webhook/router"
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

	if err := gormDB.AutoMigrate(&customermodel.Customer{}, &purchasemodel.Purchase{}, &rewardmodel.Reward{}); err != nil {
		logger.Error("auto migrate", "err", err)
		os.Exit(1)
	}
	logger.Info("schema migrated")

	// Composition root: the ONLY file that knows about every module's
	// adapters. Each module exposes ports (interfaces); main picks an
	// implementation per port and wires services to controllers.
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

	customerController := customerctl.NewController(customerService)
	purchaseController := purchasectl.NewController(purchaseService)
	rewardController := rewardctl.NewController(rewardService)
	tierController := tierctl.NewController()
	webhookController := webhookctl.NewController(customerService, purchaseService)

	// Each module registers its own routes onto the shared mux.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	customerrouter.Register(mux, customerController)
	purchaserouter.Register(mux, purchaseController)
	rewardrouter.Register(mux, rewardController)
	tierrouter.Register(mux, tierController)
	webhookrouter.Register(mux, webhookController)

	handler := requestLogger(logger)(recoverer(logger)(mux))

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

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
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
