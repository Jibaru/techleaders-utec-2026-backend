package webhook

import purchaseview "hexagonal/internal/view/purchase"

type AcceptedResponse struct {
	Status   string             `json:"status"`
	Purchase purchaseview.Response `json:"purchase"`
}
