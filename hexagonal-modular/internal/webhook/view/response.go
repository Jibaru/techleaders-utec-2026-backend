package view

import purchaseview "hexagonal-modular/internal/purchase/view"

type AcceptedResponse struct {
	Status   string             `json:"status"`
	Purchase purchaseview.Response `json:"purchase"`
}
