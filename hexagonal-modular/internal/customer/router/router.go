// Package router registers the customer module's HTTP routes onto a shared
// http.ServeMux. main.go is the only caller; it does the same for every module.
package router

import (
	"net/http"

	"hexagonal-modular/internal/customer/controller"
)

func Register(mux *http.ServeMux, c *controller.Controller) {
	mux.HandleFunc("POST /customers", c.Create)
	mux.HandleFunc("GET /customers", c.List)
	mux.HandleFunc("GET /customers/{id}", c.Get)
	mux.HandleFunc("PATCH /customers/{id}", c.Update)
	mux.HandleFunc("GET /customers/{id}/summary", c.Summary)
}
