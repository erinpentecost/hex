package hexcoord

import (
	"math"
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

const eps float64 = float64(5.0) * math.SmallestNonzeroFloat64

func closeEnough(a, b float64) bool {
	actualEpsilon := math.Abs(a - b)
	return actualEpsilon < eps
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

// Rotate should move a hex about a center point by some number of radians.
// TODO: this is probably broken because it doesn't account for the
// difference in coordinate systems.
func (h HexFractional) Rotate(center HexFractional, radians float64) HexFractional {
	v := h.Subtract(center)
	return HexFractional{
		Q: v.Q*math.Cos(radians) + v.R*math.Sin(radians),
		R: v.R*math.Cos(radians) - v.R*math.Sin(radians),
	}.Add(center)
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

// DotProduct returns the dot product.
// See https://en.wikipedia.org/wiki/Dot_product
func (h HexFractional) DotProduct(x HexFractional) float64 {
	return h.Q*x.Q + h.R*x.R
}
