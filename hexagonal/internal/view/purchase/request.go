// Package purchase holds the JSON view shapes for the purchase resource.
package purchase

type CreateRequest struct {
	AmountCents int64 `json:"amount_cents"`
}
