// Package router owns the HTTP route table.
// It wires URL patterns to controller methods and returns the resulting handler.
// Middleware (logging, recovery, etc.) is composed by the caller.
package router

import (
	"net/http"

	"mvc-coffee-loyalty/internal/controller/customer"
	"mvc-coffee-loyalty/internal/controller/purchase"
	"mvc-coffee-loyalty/internal/controller/reward"
	"mvc-coffee-loyalty/internal/controller/tier"
	"mvc-coffee-loyalty/internal/controller/webhook"
)

type Controllers struct {
	Customer *customer.Controller
	Purchase *purchase.Controller
	Reward   *reward.Controller
	Tier     *tier.Controller
	Webhook  *webhook.Controller
}

func New(c Controllers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health)

	// Customers
	mux.HandleFunc("POST /customers", c.Customer.Create)
	mux.HandleFunc("GET /customers", c.Customer.List)
	mux.HandleFunc("GET /customers/{id}", c.Customer.Get)
	mux.HandleFunc("PATCH /customers/{id}", c.Customer.Update)
	mux.HandleFunc("GET /customers/{id}/summary", c.Customer.Summary)

	// Purchases
	mux.HandleFunc("POST /customers/{id}/purchases", c.Purchase.Create)
	mux.HandleFunc("GET /customers/{id}/purchases", c.Purchase.List)
	mux.HandleFunc("POST /purchases/{id}/refund", c.Purchase.Refund)

	// Rewards
	mux.HandleFunc("POST /customers/{id}/rewards", c.Reward.Redeem)
	mux.HandleFunc("GET /customers/{id}/rewards", c.Reward.List)

	// Catalogs
	mux.HandleFunc("GET /tiers", c.Tier.List)
	mux.HandleFunc("GET /rewards/catalog", c.Reward.Catalog)

	// Webhooks
	mux.HandleFunc("POST /webhooks/payment-confirmed", c.Webhook.PaymentConfirmed)

	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
