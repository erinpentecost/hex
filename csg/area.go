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
	minR, maxR, minQ, maxQ int
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

// Equal returns true if both areas share exactly the same hexes.
func (a *Area) Equal(b *Area) bool {
	return a.CheckBounding(b) == Equals
}

// ContainsHexes returns true if the area contains all the provided hexes.
//
// If you want to determine the overlap relationship between two areas,
// use CheckBounding().
func (a *Area) ContainsHexes(hexes ...pos.Hex) bool {
	for _, k := range hexes {
		if _, ok := a.hexes[k]; !ok {
			return false
		}
	}
	return true
}

// Iterator returns an iterator on the area, returning each hex in it.
func (a *Area) Iterator() Iterator {
	hexes := make([]pos.Hex, len(a.hexes))
	i := 0
	for k := range a.hexes {
		hexes[i] = k
		i++
	}
	return &iterator{
		idx:   -1,
		hexes: hexes,
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

func bounds(p ...pos.Hex) (minR, maxR, minQ, maxQ int) {
	if len(p) == 0 {
		panic("can't get bounds of empty slice")
	}

	minR = p[0].R
	maxR = p[0].R
	minQ = p[0].Q
	maxQ = p[0].Q

	for _, point := range p[1:] {
		minR = minInt(minR, point.R)
		maxR = maxInt(maxR, point.R)

		minQ = minInt(minQ, point.Q)
		maxQ = maxInt(maxQ, point.Q)
	}
	return
}

func boundsFromMap(hexes map[pos.Hex]struct{}) (minR, maxR, minQ, maxQ int) {
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
