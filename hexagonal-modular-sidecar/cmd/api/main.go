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

	customerctl "hexagonal-modular-sidecar/internal/customer/controller"
	customermodel "hexagonal-modular-sidecar/internal/customer/model"
	customerrepo "hexagonal-modular-sidecar/internal/customer/repository"
	customercached "hexagonal-modular-sidecar/internal/customer/repository/cached"
	customergorm "hexagonal-modular-sidecar/internal/customer/repository/gorm"
	customerrouter "hexagonal-modular-sidecar/internal/customer/router"
	customersvc "hexagonal-modular-sidecar/internal/customer/service"
	purchasectl "hexagonal-modular-sidecar/internal/purchase/controller"
	purchasemodel "hexagonal-modular-sidecar/internal/purchase/model"
	purchasegorm "hexagonal-modular-sidecar/internal/purchase/repository/gorm"
	purchaserouter "hexagonal-modular-sidecar/internal/purchase/router"
	purchasesvc "hexagonal-modular-sidecar/internal/purchase/service"
	rewardctl "hexagonal-modular-sidecar/internal/reward/controller"
	rewardmodel "hexagonal-modular-sidecar/internal/reward/model"
	rewardgorm "hexagonal-modular-sidecar/internal/reward/repository/gorm"
	rewardrouter "hexagonal-modular-sidecar/internal/reward/router"
	rewardsvc "hexagonal-modular-sidecar/internal/reward/service"
	"hexagonal-modular-sidecar/internal/shared/cache"
	cachenoop "hexagonal-modular-sidecar/internal/shared/cache/noop"
	cacheredis "hexagonal-modular-sidecar/internal/shared/cache/redis"
	"hexagonal-modular-sidecar/internal/shared/config"
	"hexagonal-modular-sidecar/internal/shared/db"
	"hexagonal-modular-sidecar/internal/shared/mail"
	mailkafka "hexagonal-modular-sidecar/internal/shared/mail/kafka"
	mailnoop "hexagonal-modular-sidecar/internal/shared/mail/noop"
	mailsmtp "hexagonal-modular-sidecar/internal/shared/mail/smtp"
	repogorm "hexagonal-modular-sidecar/internal/shared/repository/gorm"
	tierctl "hexagonal-modular-sidecar/internal/tier/controller"
	tierrouter "hexagonal-modular-sidecar/internal/tier/router"
	webhookctl "hexagonal-modular-sidecar/internal/webhook/controller"
	webhookrouter "hexagonal-modular-sidecar/internal/webhook/router"
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

	// ---- Cache adapter selection ------------------------------------------
	var cacheAdapter cache.Cache
	if cfg.Redis.Addr == "" {
		cacheAdapter = cachenoop.New()
		logger.Info("redis not configured; using noop cache")
	} else {
		rc := cacheredis.New(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
		if err := rc.Ping(ctx); err != nil {
			logger.Error("redis ping failed", "addr", cfg.Redis.Addr, "err", err)
			os.Exit(1)
		}
		cacheAdapter = rc
		logger.Info("redis cache enabled", "addr", cfg.Redis.Addr)
	}

	// ---- Mail adapter selection -------------------------------------------
	// Kafka takes priority. If KAFKA_BROKERS is unset, fall back to direct
	// SMTP (or noop if SMTP_HOST is also empty). This lets the demo run with
	// or without the docker-compose stack.
	var mailer mail.Sender
	switch {
	case len(cfg.Kafka.Brokers) > 0:
		ks := mailkafka.New(cfg.Kafka.Brokers, cfg.Kafka.MailTopic)
		defer ks.Close()
		mailer = ks
		logger.Info("kafka mail sender",
			"brokers", cfg.Kafka.Brokers, "topic", cfg.Kafka.MailTopic)
	case cfg.Mail.Host == "":
		mailer = mailnoop.New()
		logger.Info("smtp & kafka not configured; using noop mail sender")
	default:
		mailer = mailsmtp.New(cfg.Mail.Host, cfg.Mail.Port, cfg.Mail.User, cfg.Mail.Password, cfg.Mail.From)
		logger.Info("smtp mail sender (no kafka)", "host", cfg.Mail.Host)
	}

	// ---- Repository wiring -----------------------------------------------
	// Customer repo is decorated with cache (cache-aside on FindByID,
	// invalidate on Update). With the noop cache adapter this is a no-op
	// wrapper and behaves identically to the bare GORM repo.
	customerGormRepo := customergorm.New(gormDB)
	purchaseRepo := purchasegorm.New(gormDB)
	rewardRepo := rewardgorm.New(gormDB)

	var customerRepo customerrepo.Repository = customercached.New(customerGormRepo, cacheAdapter)

	transactor := repogorm.NewTransactor(gormDB, customerGormRepo, purchaseRepo, rewardRepo)

	// ---- Services ---------------------------------------------------------
	customerService := customersvc.New(customerRepo, purchaseRepo, rewardRepo)
	purchaseService := purchasesvc.New(customerRepo, purchaseRepo, transactor, mailer)
	rewardService := rewardsvc.New(customerRepo, rewardRepo, transactor, mailer)

	// ---- Controllers ------------------------------------------------------
	customerController := customerctl.NewController(customerService)
	purchaseController := purchasectl.NewController(purchaseService)
	rewardController := rewardctl.NewController(rewardService)
	tierController := tierctl.NewController()
	webhookController := webhookctl.NewController(customerService, purchaseService)

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
