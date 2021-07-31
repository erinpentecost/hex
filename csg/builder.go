package csg

import (
	"github.com/erinpentecost/hexcoord/pos"
)

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
}

// NewBuilder creates a new area builder containing zero or more hexes to start with.
func NewBuilder(hexes ...pos.Hex) Builder {
	return NewArea(hexes...)
}
