package csg

import (
	"github.com/erinpentecost/hexcoord/pos"
)

func maxInt(a, k int) int {
	if a > k {
		return a
	}
	return k
}

func minInt(a, k int) int {
	if a < k {
		return a
	}
	return k
}

// BigHex returns the set of hexes that form a larger hex area
// centered around the starting hex and with the given radius.
// The order of elements returned is not set.
// A radius of 0 will return the center hex.
func BigHex(h pos.Hex, radius int) *Area {
	area := NewArea()

	for q := -1 * radius; q <= radius; q++ {
		r1 := maxInt(-1*radius, -1*(q+radius))
		r2 := minInt(radius, (-1*q)+radius)

		for r := r1; r <= r2; r++ {
			area.hexes[pos.Hex{
				Q: q,
				R: r,
			}] = exists
		}
	}

	minR, maxR, minQ, maxQ := boundsFromMap(area.hexes)
	area.minR = minR
	area.maxR = maxR
	area.minQ = minQ
	area.maxQ = maxQ
	area.boundsClean = true

	return area
}

// Rectangle returns the set of hexes that form a rectangular
// area that's a bounding box of all the supplied points.
func Rectangle(p ...pos.Hex) *Area {
	if len(p) == 0 {
		return NewArea()
	}

	minR, maxR, minQ, maxQ := bounds(p...)

	area := NewArea()
	area.minR = minR
	area.maxR = maxR
	area.minQ = minQ
	area.maxQ = maxQ
	area.boundsClean = true

	for r := minR; r <= maxR; r++ {
		rOffset := r / 2
		for q := minQ - rOffset; q <= maxQ-rOffset; q++ {
			area.hexes[pos.Hex{
				Q: q,
				R: r,
			}] = exists
		}
	}
	return area
}

// Line traces line segments along the provided points.
func Line(p ...pos.Hex) *Area {
	switch len(p) {
	case 0:
		return NewArea()
	case 1:
		return NewArea(p[0])
	case 2:
		NewArea(p[0].LineTo(p[1])...)
	}

	// get outline
	outline := NewBuilder(p...)
	last := p[0]
	for _, point := range p[1:] {
		outline = outline.Union(NewArea(last.LineTo(point)...))
		last = point
	}

	return outline.Build()
}

// Polygon returns an area that contains a polygon whose points
// are the given hexes. Order matters! Concave polygons are allowed.
func Polygon(p ...pos.Hex) *Area {
	switch len(p) {
	case 0:
		return NewArea()
	case 1:
		return NewArea(p[0])
	case 2:
		return Line(p[0], p[1])
	}

	// get line segs.
	// why not just use Line?
	// if we end up with a hex of Q-width 1,
	// we won't correctly count whether we are inside or outside.
	outlines := make([]*Area, 0)
	last := p[0]
	for _, point := range append(p[1:], p[0]) {
		outlines = append(outlines, Line(last, point))
	}

	// scanline alg
	fill := NewArea()
	minR, maxR, minQ, maxQ := bounds(p...)
	fill.minR = minR
	fill.maxR = maxR
	fill.minQ = minQ
	fill.maxQ = maxQ
	fill.boundsClean = true

	for q := minQ; q <= maxQ; q++ {
		// sorted set of points we hit
		inside := false
		for r := minR; r <= maxR; r++ {
			testHex := pos.Hex{Q: q, R: r}

			for _, outline := range outlines {
				if _, hit := outline.hexes[testHex]; hit {
					inside = !inside
					// always include the intersection hex
					fill.hexes[testHex] = exists
				}
			}

			if inside {
				fill.hexes[testHex] = exists
			}
		}
	}

	return fill
}
