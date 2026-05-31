package tier

import (
	"net/http"

	"hexagonal/internal/controller/httpx"
	"hexagonal/internal/model"
	tierview "hexagonal/internal/view/tier"
)

func (c *Controller) List(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, tierview.NewDefinitionList(model.Tiers))
}
