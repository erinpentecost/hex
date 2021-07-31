package csg

import (
	"sort"
	"strings"
	"sync"

	"github.com/erinpentecost/hexcoord/pos"
)

var (
	_ Builder = (*Area)(nil)
	_ Builder = (*AreaBuilder)(nil)
)

// exists is a dummy object to stick into map[pos.Hex]struct{}s
// to avoid allocating a bunch of small objects.
var exists = struct{}{}

// Area is a collection of hexes.
// TODO: swap for a struct that maintains parallel bounding box
//type Area map[pos.Hex]struct{}

// Area is a collection of hexes.
type Area struct {
	hexes map[pos.Hex]struct{}
	// boundsClean is true if the bounding box is ok
	boundsClean bool
	// bounding box for the area
	minR, maxR, minQ, maxQ int
}

type op func(a *Area, b *Area) *Area

// AreaBuilder allows you to use 2-dimensional constructive solid geometry techniques
// to build collections of hexes.
type AreaBuilder struct {
	a Builder
	b Builder
	o op
}

type BuildOption byte

const (
	dontClean = iota
)

// Builder can be used to build Areas.
type Builder interface {
	// Build converts a description of an Area into an actual Area.
	Build(b ...BuildOption) *Area
	// Union combines all hexes in this Area with another.
	Union(b Builder) Builder
	// Intersection returns only those hexes shared by both areas.
	Intersection(b Builder) Builder
	// Subtract returns all those hexes in the first area that are not in the second.
	Subtract(b Builder) Builder
}

// AreaFromSlice converts a slice of hexes into a set.
func AreaFromSlice(hexes []pos.Hex) *Area {
	c := make(map[pos.Hex]struct{})
	for _, k := range hexes {
		c[k] = exists
	}
	minR, maxR, minQ, maxQ := bounds(hexes...)
	return &Area{
		hexes: c,
		minR:  minR,
		maxR:  maxR,
		minQ:  minQ,
		maxQ:  maxQ,
	}
}

// NewArea creates a new area containing zero or more hexes.
func NewArea(hexes ...pos.Hex) *Area {
	c := make(map[pos.Hex]struct{})
	for _, k := range hexes {
		c[k] = exists
	}
	minR, maxR, minQ, maxQ := bounds(hexes...)
	return &Area{
		hexes: c,
		minR:  minR,
		maxR:  maxR,
		minQ:  minQ,
		maxQ:  maxQ,
	}
}

// NewBuilder creates a new area builder containing zero or more hexes to start with.
func NewBuilder(hexes ...pos.Hex) Builder {
	return NewArea(hexes...)
}

// Equal returns true if both areas share exactly the same hexes.
func (a *Area) Equal(b *Area) bool {
	if len(a.hexes) != len(b.hexes) {
		return false
	}

	// TODO: bounding box check

	for k := range a.hexes {
		if _, ok := b.hexes[k]; !ok {
			return false
		}
	}
	return true
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

func (a *Area) Build(b ...BuildOption) *Area {
	if len(b) > 0 && !a.boundsClean {
		// TODO: clean bounds
	}
	return a
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

func (ab *AreaBuilder) Union(b Builder) Builder {
	return &AreaBuilder{
		a: ab,
		b: b,
		o: unionFn,
	}
}

func (ab *AreaBuilder) Intersection(b Builder) Builder {
	return &AreaBuilder{
		a: ab,
		b: b,
		o: intersectionFn,
	}
}

func (ab *AreaBuilder) Subtract(b Builder) Builder {
	return &AreaBuilder{
		a: ab,
		b: b,
		o: subtractFn,
	}
}

func (ab *AreaBuilder) Build(b ...BuildOption) *Area {
	// Build() allows me to defer iteration until it's needed,
	// and we can also do things concurrently.

	var c *Area
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		c = ab.b.Build()
	}()

	a := ab.a.Build()

	wg.Wait()

	return ab.o(a, c)
}

// union returns all hexes in all areas.
// this operation is commutative.
func unionFn(a *Area, b *Area) *Area {
	c := make(map[pos.Hex]struct{})
	for k := range a.hexes {
		c[k] = exists
	}
	for k := range b.hexes {
		c[k] = exists
	}
	if a.boundsClean && b.boundsClean {
		return &Area{
			hexes:       c,
			boundsClean: true,
			minR:        minInt(a.minR, b.minR),
			maxR:        maxInt(a.maxR, b.maxR),
			minQ:        minInt(a.minQ, b.minQ),
			maxQ:        maxInt(a.maxQ, b.maxQ),
		}
	}
	return &Area{
		hexes: c,
	}
}

// intersectionFn returns only those hexes that are in all areas.
// this operation is commutative.
func intersectionFn(a *Area, b *Area) *Area {

	// TODO: skip by checking bounds

	c := make(map[pos.Hex]struct{})
	for k := range b.hexes {
		if _, ok := a.hexes[k]; ok {
			c[k] = exists
		}
	}
	return &Area{
		hexes: c,
	}
}

// subtractFn returns a, but with hexes shared by b removed.
func subtractFn(a *Area, b *Area) *Area {

	// TODO: skip by checking bounds

	c := make(map[pos.Hex]struct{})

	for k := range a.hexes {
		if _, ok := b.hexes[k]; !ok {
			c[k] = exists
		}
	}

	return &Area{
		hexes: c,
	}
}
