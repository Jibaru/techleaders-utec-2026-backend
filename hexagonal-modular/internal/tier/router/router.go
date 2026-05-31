// Package router registers the tier module's HTTP routes.
package router

import (
	"net/http"

	"hexagonal-modular/internal/tier/controller"
)

func Register(mux *http.ServeMux, c *controller.Controller) {
	mux.HandleFunc("GET /tiers", c.List)
}
