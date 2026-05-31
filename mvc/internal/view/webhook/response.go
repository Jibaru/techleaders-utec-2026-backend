package webhook

import purchaseview "mvc-coffee-loyalty/internal/view/purchase"

type AcceptedResponse struct {
	Status   string             `json:"status"`
	Purchase purchaseview.Response `json:"purchase"`
}
