// Package router registers the reward module's HTTP routes.
package router

import (
	"net/http"

	"hexagonal-modular-sidecar/internal/reward/controller"
)

func Register(mux *http.ServeMux, c *controller.Controller) {
	mux.HandleFunc("POST /customers/{id}/rewards", c.Redeem)
	mux.HandleFunc("GET /customers/{id}/rewards", c.List)
	mux.HandleFunc("GET /rewards/catalog", c.Catalog)
}
