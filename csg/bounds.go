package csg

// Bounding is the return type for CheckBounding.
type Bounding byte

const (
	// Distinct means areas a and b have no hexes in common.
	Distinct Bounding = iota
	// Overlap means a and b have at least one hex in common,
	//
	// a has at least one hex not in b, and
	//
	// b has at least one hex not in a.
	Overlap
	// Contains means all the hexes in b are also in a.
	Contains
	// Contains means all the hexes in a are also in b.
	ContainedBy
	// Equals means areas a and b are the same.
	Equals
)

func (a *Area) CheckBounding(b Area) Bounding {

}
