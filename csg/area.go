package csg

import (
	"fmt"
	"sort"
	"strings"

	"github.com/erinpentecost/hexcoord/internal"
	"github.com/erinpentecost/hexcoord/pos"
)

var (
	_ Builder = (*Area)(nil)
)

// exists is a dummy object to stick into map[pos.Hex]struct{}s
// to avoid allocating a bunch of small objects.
var exists = struct{}{}

// Area is a collection of hexes.
type Area struct {
	hexes map[pos.Hex]struct{}
	// boundsClean is true if the bounding box is ok.
	// this must be false for empty areas.
	boundsClean bool
	// bounding box for the area
	minR, maxR, minQ, maxQ int64
}

// NewArea creates a new area containing one or more hexes.
func NewArea(hexes ...pos.Hex) *Area {
	c := make(map[pos.Hex]struct{})
	for _, k := range hexes {
		c[k] = exists
	}
	return (&Area{
		hexes: c,
	}).ensureBounds()
}

// Slice converts the area into a slice of hexes.
//
// You can use this to marshal an area.
func (a *Area) Slice() []pos.Hex {
	hexes := make([]pos.Hex, len(a.hexes))
	i := 0
	for k := range a.hexes {
		hexes[i] = k
		i++
	}
	return hexes
}

// Size returns the number of hexes in the area.
func (a *Area) Size() int {
	return len(a.hexes)
}

// Equals returns true if both areas share exactly the same hexes.
//
// If you need more information regarding the nature of the overlap,
// use CheckBounding().
func (a *Area) Equals(b *Area) bool {
	return a.CheckBounding(b) == Equals
}

// ContainsHexes returns true if the area contains all the provided hexes.
//
// If you want to determine the overlap relationship between two areas,
// use CheckBounding(), which is more optimized for that task.
func (a *Area) ContainsHexes(hexes ...pos.Hex) bool {
	for _, k := range hexes {
		if _, ok := a.hexes[k]; !ok {
			return false
		}
	}
	return true
}

func (a *Area) String() string {
	s := []string{}
	for k := range a.hexes {
		ks := fmt.Sprintf("{\"Q\":%d,\"R\":%d}", k.Q, k.R)
		i := sort.SearchStrings(s, ks)
		s = append(s, "")
		copy(s[i+1:], s[i:])
		s[i] = ks
	}
	return "[" + strings.Join(s, ",") + "]"
}

// ensureBounds updates the bounding box if necessary.
func (a *Area) ensureBounds() *Area {
	if len(a.hexes) == 0 {
		a.boundsClean = false
		a.minR = 0
		a.maxR = 0
		a.minQ = 0
		a.maxQ = 0
		return a
	} else if a.boundsClean {
		return a
	}

	bf := boundsFinder{}

	for p := range a.hexes {
		bf.visit(&p)
	}

	return bf.applyTo(a)
}

func (a *Area) Build() *Area {
	return a.ensureBounds()
}

func (a *Area) Union(b Builder) Builder {
	return &areaBuilder{
		a:   a,
		b:   b,
		opt: union,
	}
}

func (a *Area) Intersection(b Builder) Builder {
	return &areaBuilder{
		a:   a,
		b:   b,
		opt: intersection,
	}
}

func (a *Area) Subtract(b Builder) Builder {
	return &areaBuilder{
		a:   a,
		b:   b,
		opt: subtract,
	}
}

func (a *Area) Rotate(pivot pos.Hex, direction int) Builder {
	return a.
		Transform(internal.TranslateMatrix(-1*pivot.Q, -1*pivot.R, -1*pivot.S())).
		Transform(internal.RotateMatrix(direction)).
		Transform(internal.TranslateMatrix(pivot.Q, pivot.R, pivot.S()))
}

func (a *Area) Translate(offset pos.Hex) Builder {
	return a.Transform(internal.TranslateMatrix(offset.Q, offset.R, offset.S()))
}

func (a *Area) Transform(t [4][4]int64) Builder {
	return &areaBuilder{
		a:   a,
		t:   t,
		opt: transform,
	}
}

// Bounds returns a bounding box for the area defined by two opposite-corner
// hexes. This function returns an error if the area is empty.
func (a *Area) Bounds() (minR, maxR, minQ, maxQ int64, err error) {
	a.ensureBounds()
	if !a.boundsClean {
		err = ErrEmptyArea
	}
	minR = a.minR
	maxR = a.maxR
	minQ = a.minQ
	maxQ = a.maxQ
	return
}
