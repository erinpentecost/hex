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
func BigHex(h pos.Hex, radius int) Area {
	area := NewArea()
	for q := -1 * radius; q <= radius; q++ {
		r1 := maxInt(-1*radius, -1*(q+radius))
		r2 := minInt(radius, (-1*q)+radius)

		for r := r1; r <= r2; r++ {
			area[pos.Hex{
				Q: q,
				R: r,
			}] = exists
		}
	}
	return area
}

func bounds(p ...pos.Hex) (minR, maxR, minQ, maxQ int) {
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

// Rectangle returns the set of hexes that form a rectangular
// area that's a bounding box of all the supplied points.
func Rectangle(p ...pos.Hex) Area {
	if len(p) == 0 {
		return NewArea()
	}

	minR, maxR, minQ, maxQ := bounds(p...)

	area := NewArea()

	for r := minR; r <= maxR; r++ {
		rOffset := r / 2
		for q := minQ - rOffset; q <= maxQ-rOffset; q++ {
			area[pos.Hex{
				Q: q,
				R: r,
			}] = exists
		}
	}
	return area
}

// Ring returns a set of hexes that form a ring at the given
// radius centered on the given hex.
// A radius of 0 will return the center hex.
//
// This can also be achieved by doing BigHex(h, r).Subtract(BigHex(h,r-1)).Build().
func Ring(h pos.Hex, radius int) Area {
	area := NewArea()
	if radius == 0 {
		area[h] = exists
	} else {
		ringH := h.Add(pos.Direction(4).Multiply(radius))
		for i := 0; i < 6; i++ {
			for j := 0; j < radius; j++ {
				area[ringH] = exists
				ringH = ringH.Neighbor(i)
			}
		}
	}
	return area
}

// Line traces line segments along the provided points.
func Line(p ...pos.Hex) Area {
	switch len(p) {
	case 0:
		return NewArea()
	case 1:
		return NewArea(p[0])
	case 2:
		AreaFromSlice(p[0].LineTo(p[1]))
	}

	// get outline
	outline := NewBuilder(p...)
	last := p[0]
	for _, point := range p[1:] {
		outline = outline.Union(AreaFromSlice(last.LineTo(point)))
		last = point
	}
	return outline.Build()
}

// Polygon returns an area that contains a polygon whose points
// are the given hexes. Order matters! Concave polygons are allowed.
func Polygon(p ...pos.Hex) Area {
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
	outlines := make([]Area, 0)
	last := p[0]
	for _, point := range append(p[1:], p[0]) {
		outlines = append(outlines, Line(last, point))
	}

	// scanline alg
	fill := NewArea()
	minR, maxR, minQ, maxQ := bounds(p...)

	for q := minQ; q <= maxQ; q++ {
		// sorted set of points we hit
		inside := false
		for r := minR; r <= maxR; r++ {
			testHex := pos.Hex{Q: q, R: r}

			for _, outline := range outlines {
				if _, hit := outline[testHex]; hit {
					inside = !inside
					// always include the intersection hex
					fill[testHex] = exists
				}
			}

			if inside {
				fill[testHex] = exists
			}
		}
	}

	return fill.Build()
}
