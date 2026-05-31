package tier

import (
	"net/http"

	"mvc-coffee-loyalty/internal/controller/httpx"
	"mvc-coffee-loyalty/internal/model"
	tierview "mvc-coffee-loyalty/internal/view/tier"
)

func (c *Controller) List(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, tierview.NewDefinitionList(model.Tiers))
}
