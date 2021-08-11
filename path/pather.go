package path

import "github.com/erinpentecost/hex"

// Pather contains domain knowledge for finding a path.
type Pather interface {
	// Cost indicates the move cost between a hex and one
	// of its neighbors. Higher values are less desirable.
	// Negative costs are treated as impassable.
	Cost(a hex.Hex, direction int) int

	// EstimatedCost returns the estimated cost between
	// two hexes that are not necessarily neighbors.
	// Negative costs are treated as impassable.
	EstimatedCost(a, b hex.Hex) int
}
