package csg

import (
	"sync"

	"github.com/erinpentecost/hexcoord/pos"
)

var (
	_ Builder = (*AreaBuilder)(nil)
)

type op func(a *Area, b *Area) *Area

// AreaBuilder allows you to use 2-dimensional constructive solid geometry techniques
// to build collections of hexes.
type AreaBuilder struct {
	a Builder
	b Builder
	o op
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

func (ab *AreaBuilder) Build() *Area {
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
	// we can determine a new bounding box
	// without iterating on the points if we
	// do it now
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

	if a.boundsClean && b.boundsClean && !a.mightOverlap(b) {
		return NewArea()
	}

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

	if a.boundsClean && b.boundsClean && !a.mightOverlap(b) {
		return a
	}

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
