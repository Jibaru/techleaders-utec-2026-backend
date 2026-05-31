// Package noop holds a Sender adapter that just logs the would-be email via
// slog. Useful for local development without real SMTP credentials.
package noop

import (
	"context"
	"log/slog"
)

type Sender struct{}

func New() *Sender { return &Sender{} }

func (s *Sender) Send(ctx context.Context, to, subject, body string) error {
	slog.InfoContext(ctx, "mail: would send",
		"to", to,
		"subject", subject,
		"body_len", len(body),
	)
	return nil
}
