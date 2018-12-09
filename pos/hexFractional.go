package pos

import (
	"fmt"
	"math"

	"github.com/erinpentecost/fltcmp"
)

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

// OriginFractional returns a new hex with origin (0,0) coordinates.
func OriginFractional() HexFractional {
	return HexFractional{
		Q: 0.0,
		R: 0.0,
	}
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

// ToString converts the hex to a string.
func (h HexFractional) ToString() string {
	return fmt.Sprintf("{%.3f, %.3f, %.3f}", h.Q, h.R, h.S())
}

func round(f float64) int {
	if f > 0 {
		return int(f + 0.5)
	}
	return int(f - 0.5)
}

func closeEnough(a, b float64) bool {
	return fltcmp.AlmostEqual(a, b, 5)
}

// AlmostEquals returns true when h and x are equal or close
// enough to equal for practical matters.
func (h HexFractional) AlmostEquals(x HexFractional) bool {
	return closeEnough(h.Q, x.Q) && closeEnough(h.R, x.R)
}

// Add combines two hexes.
func (h HexFractional) Add(x HexFractional) HexFractional {
	o := HexFractional{
		Q: x.Q + h.Q,
		R: x.R + h.R,
	}
	return o
}

// Subtract combines two hexes.
func (h HexFractional) Subtract(x HexFractional) HexFractional {
	o := HexFractional{
		Q: h.Q - x.Q,
		R: h.R - x.R,
	}
	return o
}

// Multiply scales a hex by a scalar value.
func (h HexFractional) Multiply(k float64) HexFractional {
	o := HexFractional{
		Q: h.Q * k,
		R: h.R * k,
	}
	return o
}

// LerpHexFractional finds a point between a and b weighted by t.
// See https://en.wikipedia.org/wiki/Linear_interpolation
func LerpHexFractional(a HexFractional, b HexFractional, t float64) HexFractional {
	return HexFractional{
		lerpFloat(a.Q, b.Q, t),
		lerpFloat(a.R, b.R, t),
	}
}

// Length gets the length of the hex to the grid origin.
// This is the Euclidean Distance.
func (h HexFractional) Length() float64 {
	return h.DistanceTo(HexFractional{0, 0})
}

// DistanceTo returns the distance between two hexes.
// This is the Euclidean Distance.
func (h HexFractional) DistanceTo(x HexFractional) float64 {
	d := h.Subtract(x)
	return math.Sqrt(d.Q*d.Q + d.R*d.R + d.Q*d.R)
}

// Normalize returns a vector that points in the same direction
// but has a length of 1.
func (h HexFractional) Normalize() HexFractional {
	return h.Multiply(1.0 / h.Length())
}

// ProjectOn projects h onto x.
// It returns a vector parallel to x.
func (h HexFractional) ProjectOn(x HexFractional) HexFractional {
	hx, hy := h.ToCartesian()
	xx, xy := x.ToCartesian()
	//a,b dot c,d == ac + bd
	dot := func(a, b, c, d float64) float64 {
		return a*c + b*d
	}

	scalar := dot(hx, hy, xx, xy) / dot(xx, xy, xx, xy)
	return HexFractionalFromCartesian(scalar*xx, scalar*xy)
}

// Rotate should move a hex about a center point counterclockwise
// by some number of radians.
func (h HexFractional) Rotate(center HexFractional, radians float64) HexFractional {

	cart := complex(h.Subtract(center).ToCartesian())

	rotation := complex(math.Cos(-radians), math.Sin(-radians))

	rotated := cart * rotation

	return HexFractionalFromCartesian(real(rotated), imag(rotated)).Add(center)
}

// AngleTo returns the angle to x in radians.
// Will always return the inner (smaller) angle.
func (h HexFractional) AngleTo(x HexFractional) float64 {
	hi := complex(h.ToCartesian())
	xi := complex(x.ToCartesian())

	// There are a lot of intermediate variables here
	// because I encountered a compiler error otherwise.
	r := xi / hi
	rr := real(r)

	rr = math.Min(math.Max(rr, -1.0), 1.0)

	return math.Acos(rr)
}

var sqrt3 float64

func init() {
	sqrt3 = math.Sqrt(3.0)
}

// ToCartesian returns the hex in Cartesian Coordinates.
func (h HexFractional) ToCartesian() (x, y float64) {
	x = sqrt3*h.Q + sqrt3*h.R/2.0
	y = 1.5 * h.R
	return
}

// HexFractionalFromCartesian returns the hex in Cartesian Coordinates.
func HexFractionalFromCartesian(x, y float64) HexFractional {
	// rotate y by 30 degrees to get R
	return HexFractional{
		Q: x*sqrt3/3.0 - y*1.0/3.0,
		R: 2.0 / 3.0 * y,
	}
}
