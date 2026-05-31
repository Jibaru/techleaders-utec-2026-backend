package webhook

import purchaseview "mvc-service-repo/internal/view/purchase"

type AcceptedResponse struct {
	Status   string             `json:"status"`
	Purchase purchaseview.Response `json:"purchase"`
}
