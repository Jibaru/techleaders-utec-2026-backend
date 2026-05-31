package reward

import (
	"net/http"

	"mvc-coffee-loyalty/internal/controller/httpx"
	"mvc-coffee-loyalty/internal/model"
	rewardview "mvc-coffee-loyalty/internal/view/reward"
)

func (c *Controller) Catalog(w http.ResponseWriter, _ *http.Request) {
	catalog := make([]rewardview.CatalogEntry, 0, len(model.RewardCosts))
	for t, cost := range model.RewardCosts {
		catalog = append(catalog, rewardview.CatalogEntry{Type: string(t), Cost: cost})
	}
	httpx.WriteJSON(w, http.StatusOK, catalog)
}
