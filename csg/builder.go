package csg

import (
	"errors"

	"github.com/erinpentecost/hexcoord/pos"
)

var ErrEmptyArea = errors.New("no boundaries for an empty area")

// Builder can be used to build Areas.
type Builder interface {
	// Build converts a description of an Area into an actual Area.
	Build() *Area
	// Union combines all hexes in this Area with another.
	Union(b Builder) Builder
	// Intersection returns only those hexes shared by both areas.
	Intersection(b Builder) Builder
	// Subtract returns all those hexes in the first area that are not in the second.
	Subtract(b Builder) Builder
	// Rotate rotates the area about some pivot some number of sides.
	Rotate(pivot pos.Hex, direction int) Builder
	// Translate adds some offset to the area.
	Translate(offste pos.Hex) Builder
	// Transform applies a transformation hex to each hex in the area.
	//
	// This doesn't infill scaling transformations!
	Transform(t [4][4]int64) Builder
}

// NewBuilder creates a new area builder containing zero or more hexes to start with.
func NewBuilder(hexes ...pos.Hex) Builder {
	return NewArea(hexes...)
}
