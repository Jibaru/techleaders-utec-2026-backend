// Package router registers the purchase module's HTTP routes.
package router

import (
	"net/http"

	"hexagonal-modular/internal/purchase/controller"
)

func Register(mux *http.ServeMux, c *controller.Controller) {
	mux.HandleFunc("POST /customers/{id}/purchases", c.Create)
	mux.HandleFunc("GET /customers/{id}/purchases", c.List)
	mux.HandleFunc("POST /purchases/{id}/refund", c.Refund)
}
