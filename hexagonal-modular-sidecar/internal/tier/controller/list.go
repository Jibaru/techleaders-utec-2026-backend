package controller

import (
	"net/http"

	"hexagonal-modular-sidecar/internal/shared/httpx"
	"hexagonal-modular-sidecar/internal/shared/model"
	tierview "hexagonal-modular-sidecar/internal/tier/view"
)

func (c *Controller) List(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, tierview.NewDefinitionList(model.Tiers))
}
