package pos

import (
	"fmt"
	"math"

	"github.com/erinpentecost/hexcoord/internal"
)

// Hex is a coordinate defined axially.
//
// [Q,R,S]
type Hex struct {
	Q int64
	R int64
}

// S is the implicit additional coordinate when using cubic coordinate system.
func (h Hex) S() int64 {
	return -1 * (h.Q + h.R)
}

// ToHexFractional returns the fractional hex that is the center of this hex.
func (h Hex) ToHexFractional() HexFractional {
	return HexFractional{
		Q: float64(h.Q),
		R: float64(h.R),
	}
}

// Origin returns a new hex with origin (0,0) coordinates.
func Origin() Hex {
	return Hex{
		Q: 0,
		R: 0,
	}
}

// Direction returns a new hex coord offset from the origin
// in the given direction, which is a number from 0 to 5, inclusive.
// Positive Q axis is in the 0 direction.
// Positive R axis is in the 5 direction.
func Direction(direction int) Hex {
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
		Q: h.Q - x.Q,
		R: h.R - x.R,
	}
	return o
}

// Multiply scales a hex by a scalar value.
func (h Hex) Multiply(k int64) Hex {
	o := Hex{
		Q: h.Q * k,
		R: h.R * k,
	}
	return o
}

func lerpFloat(a, b, t float64) float64 {
	return a*(1.0-t) + b*t
}

func lerpInt(a int64, b int64, t float64) float64 {
	return lerpFloat(float64(a), float64(b), t)
}

// LineTo returns all hexes in a line from point x to point b, inclusive.
// The order of elements is a line as you would expect.
func (h Hex) LineTo(x Hex) []Hex {
	n := h.DistanceTo(x)
	line := make([]Hex, 0)
	step := 1.0 / math.Max(float64(n), 1.0)
	for i := int64(0); i <= n; i++ {
		line = append(line, LerpHex(h, x, step*float64(i)))
	}
	return line
}

// LerpHex finds a point between a and b weighted by t.
// See https://en.wikipedia.org/wiki/Linear_interpolation
func LerpHex(a Hex, b Hex, t float64) Hex {
	hf := HexFractional{
		lerpInt(a.Q, b.Q, t),
		lerpInt(a.R, b.R, t),
	}
	return hf.ToHex()
}

func absInt(k int64) int64 {
	if k > 0 {
		return k
	}

	return -1 * k
}

// Length gets the length of the hex to the grid origin.
//
// This is the Manhattan Distance.
func (h Hex) Length() int64 {
	return (absInt(h.Q) + absInt(h.R) + absInt(h.S())) / 2
}

// DistanceTo returns the distance between two hexes.
//
// This is the Manhattan Distance.
func (h Hex) DistanceTo(x Hex) int64 {
	return h.Subtract(x).Length()
}

// Center returns the hex at the center of mass of the given points.
func Center(h ...Hex) HexFractional {
	if len(h) == 0 {
		return OriginFractional()
	}
	c := h[0]
	for _, e := range h[1:] {
		c = c.Add(e)
	}
	cf := c.ToHexFractional()
	return cf.Multiply(1.0 / float64(len(h)))
}

// Neighbor returns the neighbor in the given directon.
func (h Hex) Neighbor(direction int) Hex {
	d := Direction(direction)
	return h.Add(d)
}

// Neighbors returns the neighbors.
func (h Hex) Neighbors() []Hex {
	n := make([]Hex, 7)
	for i := 0; i <= 6; i++ {
		n[i] = h.Neighbor(i)
	}
	return n
}

// Vertex returns one point on the Hex, which is the point
// between this hex, it's Neighbor(direction), and Neighbor(direction+1).
func (h Hex) Vertex(direction int) HexFractional {
	// TODO: optimize this
	return Center(h, h.Neighbor(direction), h.Neighbor(direction+1))
}

// Transform applies a matrix transformation on the hex.
//
// Translation by tr,tq,ts:
//
// [[1,0,0,tr]
//
// [0,1,0,tq]
//
// [0,0,1,ts]
//
// [0,0,0,1]] // homogenous coords. ignored.
func (h Hex) Transform(t [4][4]int64) Hex {
	p := Hex{
		Q: t[0][0]*h.Q + t[0][1]*h.R + t[0][2]*h.S() + t[0][3],
		R: t[1][0]*h.Q + t[1][1]*h.R + t[1][2]*h.S() + t[1][3],
	}
	/*
		// No need to transform S since it's a derived field.
		s := t[2][0]*h.Q + t[2][1]*h.R + t[2][2]*h.S() + t[2][3]
		if p.S() != s {
			panic("transformation matrix is bad")
		}
	*/
	return p
}

func (h Hex) Rotate(pivot Hex, direction int) Hex {
	d := BoundFacing(direction)

	if d == 0 {
		return h
	}

	if (pivot == Hex{}) {
		return h.Transform(internal.RotationMatrixes[d])
	}
	return h.Subtract(pivot).Transform(internal.RotationMatrixes[d]).Add(pivot)
}

// BoundFacing maps the whole number set to 0-5.
func BoundFacing(facing int) int {
	return internal.BoundFacing(facing)
}

// ToString converts the hex to a string.
func (h Hex) String() string {
	return fmt.Sprintf("{%v, %v, %v}", h.Q, h.R, h.S())
}
