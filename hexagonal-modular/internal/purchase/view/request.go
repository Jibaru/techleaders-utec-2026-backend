// Package purchase holds the JSON view shapes for the purchase resource.
package view

type CreateRequest struct {
	AmountCents int64 `json:"amount_cents"`
}
