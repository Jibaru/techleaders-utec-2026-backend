package tier

import (
	"net/http"

	"mvc-service-repo/internal/controller/httpx"
	"mvc-service-repo/internal/model"
	tierview "mvc-service-repo/internal/view/tier"
)

func (c *Controller) List(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, tierview.NewDefinitionList(model.Tiers))
}
