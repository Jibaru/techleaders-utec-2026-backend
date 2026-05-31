package reward

import (
	"time"

	"github.com/google/uuid"

	"mvc-service-repo/internal/model"
	tierview "mvc-service-repo/internal/view/tier"
)

type Response struct {
	ID          uuid.UUID `json:"id"`
	CustomerID  uuid.UUID `json:"customer_id"`
	Type        string    `json:"type"`
	PointsSpent int       `json:"points_spent"`
	CreatedAt   time.Time `json:"created_at"`
}

type RedeemedResponse struct {
	Reward          Response       `json:"reward"`
	RemainingPoints int            `json:"remaining_points"`
	Tier            tierview.Brief `json:"tier"`
}

type CatalogEntry struct {
	Type string `json:"type"`
	Cost int    `json:"cost"`
}

func NewResponse(r model.Reward) Response {
	return Response{
		ID:          r.ID,
		CustomerID:  r.CustomerID,
		Type:        string(r.Type),
		PointsSpent: r.PointsSpent,
		CreatedAt:   r.CreatedAt,
	}
}
