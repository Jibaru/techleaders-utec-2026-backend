// Package webhook holds the JSON view shapes for inbound webhook payloads.
package view

import "time"

type PaymentConfirmed struct {
	ExternalPaymentID string    `json:"external_payment_id"`
	CustomerEmail     string    `json:"customer_email"`
	AmountCents       int64     `json:"amount_cents"`
	OccurredAt        time.Time `json:"occurred_at"`
}
