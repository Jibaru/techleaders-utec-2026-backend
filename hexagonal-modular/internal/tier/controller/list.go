package controller

import (
	"net/http"

	"hexagonal-modular/internal/shared/httpx"
	"hexagonal-modular/internal/shared/model"
	tierview "hexagonal-modular/internal/tier/view"
)

func (c *Controller) List(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, tierview.NewDefinitionList(model.Tiers))
}
