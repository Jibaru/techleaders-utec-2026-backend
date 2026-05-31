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

	"mvc-coffee-loyalty/internal/config"
	"mvc-coffee-loyalty/internal/controller/customer"
	"mvc-coffee-loyalty/internal/controller/purchase"
	"mvc-coffee-loyalty/internal/controller/reward"
	"mvc-coffee-loyalty/internal/controller/tier"
	"mvc-coffee-loyalty/internal/controller/webhook"
	"mvc-coffee-loyalty/internal/db"
	"mvc-coffee-loyalty/internal/mail"
	"mvc-coffee-loyalty/internal/model"
	"mvc-coffee-loyalty/internal/router"
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

	mailer := mail.NewSender(cfg.Mail.Host, cfg.Mail.Port, cfg.Mail.User, cfg.Mail.Password, cfg.Mail.From)
	if cfg.Mail.Host == "" {
		logger.Info("smtp not configured; emails will be logged but not sent")
	} else {
		logger.Info("smtp configured", "host", cfg.Mail.Host, "from", cfg.Mail.From)
	}

	handler := router.New(router.Controllers{
		Customer: customer.NewController(gormDB),
		Purchase: purchase.NewController(gormDB, mailer),
		Reward:   reward.NewController(gormDB, mailer),
		Tier:     tier.NewController(),
		Webhook:  webhook.NewController(gormDB, mailer),
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
