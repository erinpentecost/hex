package area

import (
	"github.com/erinpentecost/hex"
	"github.com/erinpentecost/hex/internal"
)

func maxInt(a, k int64) int64 {
	if a > k {
		return a
	}
	return k
}

func minInt(a, k int64) int64 {
	if a < k {
		return a
	}
	return k
}

// BigHex returns the set of hexes that form a larger hex area
// centered around the starting hex and with the given radius.
// The order of elements returned is not set.
// A radius of 0 will return the center hex.
func BigHex(center hex.Hex, radius int64) *Area {
	area := NewArea()
	bf := boundsFinder{}
	for q := -1 * radius; q <= radius; q++ {
		r1 := maxInt(-1*radius, -1*(q+radius))
		r2 := minInt(radius, (-1*q)+radius)

		for r := r1; r <= r2; r++ {
			h := hex.Hex{
				Q: q + center.Q,
				R: r + center.R,
			}
			area.hexes[h] = exists
			bf.visit(&h)
		}
	}

	return bf.applyTo(area)
}

// Circle draws a circle. At small radiuses, this is just like BigHex.
func Circle(center hex.Hex, radius int64) *Area {

	// find some bounding box that contains the circle
	edgePoint := hex.HexFractionalFromCartesian(0, float64(radius+1)).ToHex()
	p := []hex.Hex{}
	for i := 0; i < 6; i++ {
		p = append(p, edgePoint.Rotate(hex.Origin(), i))
	}
	bounds := NewArea(p...)

	// check every hex in the box and test if the hex is in the circle
	area := NewArea()
	rs := float64(radius * radius)
	for q := bounds.minQ; q <= bounds.maxQ; q++ {
		for r := bounds.minR; r <= bounds.maxR; r++ {
			x, y := hex.HexFractional{Q: float64(q), R: float64(r)}.ToCartesian()
			dist := x*x + y*y
			if dist <= rs || internal.CloseEnough(dist, rs) {
				area.hexes[hex.Hex{Q: q, R: r}] = exists
			}
		}
	}

	// move circle center
	return area.Translate(center).Build()
}

// Rectangle returns the set of hexes that form a rectangular
// area that's a bounding box of all the supplied points.
func Rectangle(p ...hex.Hex) *Area {
	if len(p) == 0 {
		return NewArea()
	}

	area := NewArea(p...)

	for r := area.minR; r <= area.maxR; r++ {
		rOffset := r / 2
		for q := area.minQ - rOffset; q <= area.maxQ-rOffset; q++ {
			area.hexes[hex.Hex{
				Q: q,
				R: r,
			}] = exists
		}
	}
	return area
}

// Line traces line segments along the provided points.
func Line(p ...hex.Hex) *Area {
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
func Polygon(p ...hex.Hex) *Area {
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

	return fill(outlines, p)
}

func fill(edges []*Area, vertices []hex.Hex) *Area {
	// scanline alg
	f := NewArea(vertices...)

	for q := f.minQ; q <= f.maxQ; q++ {
		// sorted set of points we hit
		inside := false
		for r := f.minR; r <= f.maxR; r++ {
			testHex := hex.Hex{Q: q, R: r}

			for _, outline := range edges {
				if _, hit := outline.hexes[testHex]; hit {
					inside = !inside
					// always include the intersection hex
					f.hexes[testHex] = exists
				}
			}

			if inside {
				f.hexes[testHex] = exists
			}
		}
	}

	return f
}
