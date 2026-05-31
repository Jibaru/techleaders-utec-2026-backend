// Package router registers the webhook module's HTTP routes.
package router

import (
	"net/http"

	"hexagonal-modular-sidecar/internal/webhook/controller"
)

func Register(mux *http.ServeMux, c *controller.Controller) {
	mux.HandleFunc("POST /webhooks/payment-confirmed", c.PaymentConfirmed)
}
