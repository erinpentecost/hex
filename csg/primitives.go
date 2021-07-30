package csg

import "github.com/erinpentecost/hexcoord/pos"

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

// Rectangle returns the set of hexes that form a rectangular
// area from the given hex to another hex representing an opposite corner.
func Rectangle(h pos.Hex, opposite pos.Hex) Area {
	area := NewArea()

	minR := minInt(h.R, opposite.R)
	maxR := maxInt(h.R, opposite.R)

	minQ := minInt(h.Q, opposite.Q)
	maxQ := maxInt(h.Q, opposite.Q)
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

// Line returns all hexes in a line from point a to point b, inclusive.
func Line(a pos.Hex, b pos.Hex) Area {
	return AreaFromSlice(a.LineTo(b))
}
