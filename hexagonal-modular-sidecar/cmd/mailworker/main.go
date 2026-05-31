// Command mailworker consumes mail messages from Kafka and delivers them via
// SMTP. It is the consumer-side counterpart to internal/shared/mail/kafka.Sender.
//
// Running this in a separate process means the API can ack the user's request
// the moment the mail is enqueued, while real SMTP latency / outages are
// absorbed by the queue.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"hexagonal-modular-sidecar/internal/shared/config"
	mailkafka "hexagonal-modular-sidecar/internal/shared/mail/kafka"
	mailsmtp "hexagonal-modular-sidecar/internal/shared/mail/smtp"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "err", err)
		os.Exit(1)
	}
	if len(cfg.Kafka.Brokers) == 0 {
		logger.Error("KAFKA_BROKERS is required for mailworker")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	smtpSender := mailsmtp.New(cfg.Mail.Host, cfg.Mail.Port, cfg.Mail.User, cfg.Mail.Password, cfg.Mail.From)
	if cfg.Mail.Host == "" {
		logger.Warn("SMTP_HOST is empty; mailworker will log payloads but not send")
	}

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        cfg.Kafka.Brokers,
		Topic:          cfg.Kafka.MailTopic,
		GroupID:        "mail-worker",
		MinBytes:       1,
		MaxBytes:       10 << 20, // 10 MiB
		MaxWait:        500 * time.Millisecond,
		CommitInterval: 0, // commit synchronously after each successful send
	})
	defer reader.Close()

	logger.Info("mailworker listening",
		"brokers", cfg.Kafka.Brokers,
		"topic", cfg.Kafka.MailTopic,
	)

	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				logger.Info("shutdown signal received")
				return
			}
			logger.Error("kafka fetch", "err", err)
			time.Sleep(time.Second)
			continue
		}

		var msg mailkafka.Message
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			logger.Error("decode message; skipping",
				"err", err, "offset", m.Offset, "partition", m.Partition,
			)
			_ = reader.CommitMessages(ctx, m)
			continue
		}

		if cfg.Mail.Host == "" {
			logger.Info("would send (smtp not configured)",
				"to", msg.To, "subject", msg.Subject,
			)
		} else if err := smtpSender.Send(ctx, msg.To, msg.Subject, msg.Body); err != nil {
			// Don't commit — the message will be re-delivered after the
			// consumer group's session times out. In production you'd add a
			// DLQ and bounded retries; for the demo, "keep trying" is fine.
			logger.Error("smtp send", "err", err, "to", msg.To)
			continue
		} else {
			logger.Info("sent", "to", msg.To, "subject", msg.Subject)
		}

		if err := reader.CommitMessages(ctx, m); err != nil {
			logger.Error("kafka commit", "err", err)
		}
	}
}
