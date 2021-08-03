package csg

import (
	"sync"

	"github.com/erinpentecost/hexcoord/internal"
	"github.com/erinpentecost/hexcoord/pos"
)

var (
	_ Builder = (*areaBuilder)(nil)
)

type operation byte

const (
	union operation = iota
	intersection
	subtract
	transform
)

// areaBuilder allows you to use 2-dimensional constructive solid geometry techniques
// to build collections of hexes.
type areaBuilder struct {
	a   Builder
	b   Builder
	t   [4][4]int64
	opt operation
}

func (ab *areaBuilder) Union(b Builder) Builder {
	// TODO optimization: rotate subtree
	return &areaBuilder{
		a:   ab,
		b:   b,
		opt: union,
	}
}

func (ab *areaBuilder) Intersection(b Builder) Builder {
	// TODO optimization: rotate subtree
	return &areaBuilder{
		a:   ab,
		b:   b,
		opt: intersection,
	}
}

func (ab *areaBuilder) Subtract(b Builder) Builder {
	return &areaBuilder{
		a:   ab,
		b:   b,
		opt: subtract,
	}
}

func (ab *areaBuilder) Rotate(pivot pos.Hex, direction int) Builder {
	return ab.
		Transform(internal.TranslateMatrix(-1*pivot.Q, -1*pivot.R)).
		Transform(internal.RotateMatrix(direction)).
		Transform(internal.TranslateMatrix(pivot.Q, pivot.R))
}

func (ab *areaBuilder) Translate(offset pos.Hex) Builder {
	return ab.Transform(internal.TranslateMatrix(offset.Q, offset.R))
}

// Transform applies a transformation matrix to all hexes in ab.
func (ab *areaBuilder) Transform(t [4][4]int64) Builder {
	return &areaBuilder{
		a:   ab,
		t:   t,
		opt: transform,
	}
}

func (ab *areaBuilder) Build() *Area {
	if ab.opt == transform {
		// TODO optimization: combine transforms then apply
		// instead of blindly building
		a := ab.a.Build()

		if len(a.hexes) == 0 {
			return a
		}

		// apply transform to all hexes
		// TODO optimization: calculate bounds
		bf := boundsFinder{}
		out := NewArea()
		for k := range a.hexes {
			h := k.Transform(ab.t)
			out.hexes[h] = exists
			bf.visit(&h)
		}

		return bf.applyTo(out)
	}

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

	switch ab.opt {
	case union:
		return unionFn(a, c)
	case intersection:
		return intersectionFn(a, c)
	case subtract:
		return subtractFn(a, c)
	}
	panic("unsupported operation")
}

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
