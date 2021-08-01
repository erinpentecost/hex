package csg

import (
	"sort"
	"strings"

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

// NewArea creates a new area containing zero or more hexes.
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

// Equal returns true if both areas share exactly the same hexes.
//
// If you need more information regarding the nature of the overlap,
// use CheckBounding().
func (a *Area) Equal(b *Area) bool {
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

// Iterator returns an iterator on the area, returning each hex in it.
//
// If you want all the hexes, use Slice().
func (a *Area) Iterator() Iterator {
	return &iterator{
		idx:   -1,
		hexes: a.Slice(),
	}
}

func (a *Area) String() string {
	s := []string{}
	for k := range a.hexes {
		ks := k.String()
		i := sort.SearchStrings(s, ks)
		s = append(s, "")
		copy(s[i+1:], s[i:])
		s[i] = ks
	}
	return "Area: {" + strings.Join(s, " ") + "}"
}

// ensureBounds updates the bounding box if necessary.
func (a *Area) ensureBounds() *Area {
	if len(a.hexes) == 0 {
		a.boundsClean = false
	} else if !a.boundsClean {
		minR, maxR, minQ, maxQ := boundsFromMap(a.hexes)
		a.minR = minR
		a.maxR = maxR
		a.minQ = minQ
		a.maxQ = maxQ
		a.boundsClean = true
	}
	return a
}

func (a *Area) Build() *Area {
	return a.ensureBounds()
}

func (a *Area) Union(b Builder) Builder {
	return &AreaBuilder{
		a: a,
		b: b,
		o: unionFn,
	}
}

func (a *Area) Intersection(b Builder) Builder {
	return &AreaBuilder{
		a: a,
		b: b,
		o: intersectionFn,
	}
}

func (a *Area) Subtract(b Builder) Builder {
	return &AreaBuilder{
		a: a,
		b: b,
		o: subtractFn,
	}
}

func (a *Area) Bounds() (minR, maxR, minQ, maxQ int64) {
	if len(a.hexes) == 0 {
		panic("can't determine bounds for an empty area")
	}
	a.ensureBounds()
	minR = a.minR
	maxR = a.maxR
	minQ = a.minQ
	maxQ = a.maxQ
	return
}

type Iterator interface {
	Next() *pos.Hex
}

type iterator struct {
	// idx should be init-ed as -1
	idx   int
	hexes []pos.Hex
}

func (i *iterator) Next() *pos.Hex {
	i.idx += 1
	if i.idx < len(i.hexes) {
		return &i.hexes[i.idx]
	}
	return nil
}

func boundsFromMap(hexes map[pos.Hex]struct{}) (minR, maxR, minQ, maxQ int64) {
	if len(hexes) == 0 {
		panic("can't get bounds of empty map")
	}

	for p := range hexes {
		minR = p.R
		maxR = p.R
		minQ = p.Q
		maxQ = p.Q
	}

	for p := range hexes {
		minR = minInt(minR, p.R)
		maxR = maxInt(maxR, p.R)

		minQ = minInt(minQ, p.Q)
		maxQ = maxInt(maxQ, p.Q)
	}
	return
}
