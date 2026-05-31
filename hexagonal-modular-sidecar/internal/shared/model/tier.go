package model

type Tier struct {
	Name       string
	MinPoints  int
	Multiplier float64
}

var Tiers = []Tier{
	{Name: "Bronze", MinPoints: 0, Multiplier: 1.00},
	{Name: "Silver", MinPoints: 500, Multiplier: 1.25},
	{Name: "Gold", MinPoints: 2000, Multiplier: 1.50},
}

// TierForPoints walks the tier ladder to find which tier a point balance falls into.
func TierForPoints(points int) Tier {
	current := Tiers[0]
	for _, t := range Tiers {
		if points >= t.MinPoints {
			current = t
		}
	}
	return current
}

// NextTierForPoints returns the next tier above the customer's current one,
// or nil if they are already at the top of the ladder.
func NextTierForPoints(points int) *Tier {
	current := TierForPoints(points)
	for i, t := range Tiers {
		if t.Name == current.Name && i+1 < len(Tiers) {
			next := Tiers[i+1]
			return &next
		}
	}
	return nil
}
