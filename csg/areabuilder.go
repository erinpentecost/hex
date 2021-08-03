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
	noop
)

func (o operation) String() string {
	switch o {
	case union:
		return "u"
	case intersection:
		return "i"
	case subtract:
		return "s"
	case transform:
		return "t"
	case noop:
		return "n"
	default:
		return "?"
	}
}

// areaBuilder allows you to use 2-dimensional constructive solid geometry techniques
// to build collections of hexes.
// this is a node of a tree with a and b being the two child nodes
type areaBuilder struct {
	left  Builder
	right Builder
	t     [4][4]int64
	opt   operation
}

func height(b Builder) int64 {
	if bb, ok := b.(*areaBuilder); ok {
		return maxInt(height(bb.left), height(bb.right)) + 1
	}
	return 1
}

func (ab *areaBuilder) Union(b Builder) Builder {
	// at this point, we are inserting a new root node (that is returned)
	// whose child a is the old root (ab) and child b is the new builder we are adding.
	// this has an effect of making the tree really unbalanced when doing long chains
	// of operations. in these cases, we get a long chain of nodes along the `a` path
	// but only 1 node on the `b` path.

	// We can optimize this since unions are associative operations.

	// Normally, we make a new node and stick ab as a child and b as another.
	// But if those children have different heights AND
	// both those children are also unions, THEN
	// we can do a rotation insertion.

	// optimization: rotate subtree
	// bad (big b): union(a, union(b, union(c, d)))
	// bad (big a): union(union(union(a, b), c), d)
	// good: union(union(a, b), union(c,d))

	return &areaBuilder{
		left:  ab,
		right: b,
		opt:   union,
	}
}

func (ab *areaBuilder) Intersection(b Builder) Builder {

	// We can optimize interstions in the same way as unions since insertions
	// are also associative.

	return &areaBuilder{
		left:  ab,
		right: b,
		opt:   intersection,
	}
}

func (ab *areaBuilder) Subtract(b Builder) Builder {

	// Subtractions can't be optimized.

	return &areaBuilder{
		left:  ab,
		right: b,
		opt:   subtract,
	}
}

func (ab *areaBuilder) Rotate(pivot pos.Hex, direction int) Builder {
	return ab.
		Transform(internal.TranslateMatrix(-1*pivot.Q, -1*pivot.R, -1*pivot.S())).
		Transform(internal.RotateMatrix(direction)).
		Transform(internal.TranslateMatrix(pivot.Q, pivot.R, pivot.S()))
}

func (ab *areaBuilder) Translate(offset pos.Hex) Builder {
	return ab.Transform(internal.TranslateMatrix(offset.Q, offset.R, offset.S()))
}

// Transform applies a transformation matrix to all hexes in ab.
func (ab *areaBuilder) Transform(t [4][4]int64) Builder {
	// if we are chaining transforms, combine them.
	if ab.opt == transform {
		// ab.t is applied first, then t.
		ab.t = internal.MatrixMultiply(t, ab.t)
		return ab
	}
	return &areaBuilder{
		left: ab,
		t:    t,
		opt:  transform,
	}
}

func (ab *areaBuilder) Build() *Area {
	if ab.opt == noop {
		return ab.left.Build()
	}

	if ab.opt == transform {
		a := ab.left.Build()

		if len(a.hexes) == 0 {
			return a
		}

		// apply transform to all hexes
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
		c = ab.right.Build()
	}()

	a := ab.left.Build()

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
