package hexcoord

import (
	"math"
)

// Hex is a coordinate defined axially.
type Hex struct {
	Q int
	R int
}

// S is the implicit additional coordinate when using cubic coordinate system.
func (h Hex) S() int {
	return -1 * (h.Q + h.R)
}

// ToHexFractional returns the fractional hex that is the center of this hex.
func (h Hex) ToHexFractional() HexFractional {
	return HexFractional{
		Q: float64(h.Q),
		R: float64(h.R),
	}
}

// HexFractional is fractional hex coordinates in
// cubic coordinate system.
type HexFractional struct {
	Q float64
	R float64
}

// S is the implicit additional coordinate when using cubic coordinate system.
func (h HexFractional) S() float64 {
	return -1 * (h.Q + h.R)
}

// ToHex takes in fractional hex coordinates in
// cubic coordinates and rounds them to the nearest
// actual hex coordinate. This is all in normal coordinate
// space, not screen space.
func (h HexFractional) ToHex() Hex {
	q := round(h.Q)
	r := round(h.R)
	s := round(h.S())

	qd := math.Abs(float64(q) - h.Q)
	rd := math.Abs(float64(r) - h.R)
	sd := math.Abs(float64(s) - h.S())

	if qd > rd && qd > sd {
		q = -r - s
	} else if rd > sd {
		r = -q - s
	}

	return Hex{
		Q: q,
		R: r,
	}
}

func round(f float64) int {
	if f > 0 {
		return int(f + 0.5)
	}
	return int(f - 0.5)
}

// HexOrigin returns a new hex with origin (0,0) coordinates.
func HexOrigin() Hex {
	return Hex{
		Q: 0,
		R: 0,
	}
}

// HexDirection returns a new hex coord offset from the origin
// in the given direction, which is a number from 0 to 5, inclusive
func HexDirection(direction int) Hex {
	d := BoundFacing(direction)

	switch d {
	case 0:
		return Hex{
			Q: 1,
			R: 0,
		}
	case 1:
		return Hex{
			Q: 1,
			R: -1,
		}
	case 2:
		return Hex{
			Q: 0,
			R: -1,
		}
	case 3:
		return Hex{
			Q: -1,
			R: 0,
		}
	case 4:
		return Hex{
			Q: -1,
			R: 1,
		}
	case 5:
		return Hex{
			Q: 0,
			R: 1,
		}
	}
	panic("should never get here.")
}

// Add combines two hexes.
func (h Hex) Add(x Hex) Hex {
	o := Hex{
		Q: x.Q + h.Q,
		R: x.R + h.R,
	}
	return o
}

// Subtract combines two hexes.
func (h Hex) Subtract(x Hex) Hex {
	o := Hex{
		Q: x.Q - h.Q,
		R: x.R - h.R,
	}
	return o
}

// Multiply scales a hex by a scalar value.
func (h Hex) Multiply(k int) Hex {
	o := Hex{
		Q: h.Q * k,
		R: h.R * k,
	}
	return o
}

func lerpFloat(a, b, t float64) float64 {
	return a*(1.0-t) + b*t
}

func lerpInt(a int, b int, t float64) float64 {
	return lerpFloat(float64(a), float64(b), t)
}

func lerpHexFractional(a Hex, b Hex, t float64) HexFractional {
	return HexFractional{
		lerpInt(a.Q, b.Q, t),
		lerpInt(a.R, b.R, t),
	}
}

func lerpHex(a Hex, b Hex, t float64) Hex {
	hf := HexFractional{
		lerpInt(a.Q, b.Q, t),
		lerpInt(a.R, b.R, t),
	}
	return hf.ToHex()
}

func absInt(k int) int {
	if k > 0 {
		return k
	}

	return -1 * k
}

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

// Length gets the length of the hex to the grid origin.
func (h Hex) Length() int {
	return (absInt(h.Q) + absInt(h.R) + absInt(h.S())) / 2
}

// DistanceTo returns the distance between two hexes.
func (h Hex) DistanceTo(x Hex) int {
	return h.Subtract(x).Length()
}

// Neighbor returns the neighbor in the given directon.
func (h Hex) Neighbor(direction int) Hex {
	d := HexDirection(direction)
	return h.Add(d)
}

// Rotate rotates a hex X times clockwise.
// The value can be negative.
// The number of degrees rotated is 60*direction
// This could be made faster.
func (h Hex) Rotate(pivot Hex, direction int) Hex {
	d := BoundFacing(direction)

	if d == 0 {
		return h
	}

	rotated := Hex{
		Q: -h.S(),
		R: -h.Q,
	}

	return rotated.Rotate(pivot, d-1)
}

// BoundFacing maps the whole number set to 0-5.
func BoundFacing(facing int) int {
	d := facing % 6
	if d < 0 {
		d = d + 6
	}
	return d
}
