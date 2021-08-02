package csg

import (
	"github.com/erinpentecost/hexcoord/pos"
)

var (
	_ Builder = (*areaBuilderUnaryOp)(nil)
)

type unOp func(a *Area) *Area

// areaBuilderUnaryOp allows you to use 2-dimensional constructive solid geometry techniques
// to build collections of hexes.
type areaBuilderUnaryOp struct {
	a Builder
	o unOp
}

func (ab *areaBuilderUnaryOp) Union(b Builder) Builder {
	return &areaBuilderBinaryOp{
		a: ab,
		b: b,
		o: unionFn,
	}
}

func (ab *areaBuilderUnaryOp) Intersection(b Builder) Builder {
	return &areaBuilderBinaryOp{
		a: ab,
		b: b,
		o: intersectionFn,
	}
}

func (ab *areaBuilderUnaryOp) Subtract(b Builder) Builder {
	return &areaBuilderBinaryOp{
		a: ab,
		b: b,
		o: subtractFn,
	}
}

func (ab *areaBuilderUnaryOp) Build() *Area {
	return ab.o(ab.a.Build())
}

func (ab *areaBuilderUnaryOp) Rotate(pivot pos.Hex, direction int) Builder {
	return &areaBuilderUnaryOp{
		a: ab,
		o: rotateFn(pivot, direction),
	}
}

func (ab *areaBuilderUnaryOp) Translate(offset pos.Hex) Builder {
	return &areaBuilderUnaryOp{
		a: ab,
		o: translateFn(offset),
	}
}

func rotateFn(pivot pos.Hex, direction int) unOp {
	return func(a *Area) *Area {
		if len(a.hexes) == 0 {
			return a
		}

		c := make(map[pos.Hex]struct{})
		var minR, maxR, minQ, maxQ int64

		for k := range a.hexes {
			p := k.Rotate(pivot, direction)
			c[p] = exists

			minR = minInt(minR, p.R)
			maxR = maxInt(maxR, p.R)

			minQ = minInt(minQ, p.Q)
			maxQ = maxInt(maxQ, p.Q)
		}

		return &Area{
			hexes:       c,
			boundsClean: true,
			minR:        minR,
			maxR:        maxR,
			minQ:        minQ,
			maxQ:        maxQ,
		}
	}
}

func translateFn(offset pos.Hex) unOp {
	return func(a *Area) *Area {
		if len(a.hexes) == 0 {
			return a
		}

		c := make(map[pos.Hex]struct{})
		var minR, maxR, minQ, maxQ int64

		for k := range a.hexes {
			p := k.Add(offset)
			c[p] = exists

			minR = minInt(minR, p.R)
			maxR = maxInt(maxR, p.R)

			minQ = minInt(minQ, p.Q)
			maxQ = maxInt(maxQ, p.Q)
		}

		return &Area{
			hexes:       c,
			boundsClean: true,
			minR:        minR,
			maxR:        maxR,
			minQ:        minQ,
			maxQ:        maxQ,
		}
	}
}
