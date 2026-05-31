package controller

import (
	"net/http"

	"hexagonal-modular/internal/shared/httpx"
	"hexagonal-modular/internal/reward/model"
	rewardview "hexagonal-modular/internal/reward/view"
)

func (c *Controller) Catalog(w http.ResponseWriter, _ *http.Request) {
	catalog := make([]rewardview.CatalogEntry, 0, len(model.RewardCosts))
	for t, cost := range model.RewardCosts {
		catalog = append(catalog, rewardview.CatalogEntry{Type: string(t), Cost: cost})
	}
	httpx.WriteJSON(w, http.StatusOK, catalog)
}
