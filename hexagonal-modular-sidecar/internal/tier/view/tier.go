// Package tier holds the JSON view shapes for the loyalty tier ladder.
package view

import "hexagonal-modular-sidecar/internal/shared/model"

// Brief is the embedded tier representation used inside other responses
// (customer, reward redemption, etc.).
type Brief struct {
	Name       string  `json:"name"`
	Multiplier float64 `json:"multiplier"`
}

// Definition is the full tier representation returned by the catalog endpoint.
type Definition struct {
	Name       string  `json:"name"`
	MinPoints  int     `json:"min_points"`
	Multiplier float64 `json:"multiplier"`
}

func NewBrief(t model.Tier) Brief {
	return Brief{Name: t.Name, Multiplier: t.Multiplier}
}

func NewDefinitionList(ts []model.Tier) []Definition {
	out := make([]Definition, 0, len(ts))
	for _, t := range ts {
		out = append(out, Definition{
			Name:       t.Name,
			MinPoints:  t.MinPoints,
			Multiplier: t.Multiplier,
		})
	}
	return out
}
