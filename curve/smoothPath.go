package curve

import (
	"fmt"
	"math"

	"github.com/erinpentecost/hexcoord/pos"
)

// CircularArc defines a circular arc in three vectors.
type CircularArc struct {
	// I is the initial point.
	I pos.HexFractional
	// T is the tangent unit vector at the initial point.
	T pos.HexFractional
	// E is the end point.
	E pos.HexFractional
}

// ToString converts the arc to a string.
func (ca CircularArc) ToString() string {
	return fmt.Sprintf("{I: %v, T: %v, E: %v}", ca.I.ToString(), ca.T.ToString(), ca.E.ToString())
}

// SmoothPath takes as input a slice of connected Hexes.
// As output, it will return a function that describes a
// series of connected circular arcs that pass through all
// original hexes and no additional hexes.
// Arcs with infinite radius (straight lines) are allowed
// so long as it remains G1 continuous.
// It will also return a vector tangent to the movement arc.
// 0.0f is the start position, and 1.0f is the end position.
// Unlike other functions in this package, it assumes hexes
// are regular.
// This function can be used to generate smooth movement.
func SmoothPath(done <-chan interface{}, ti, te, path []pos.HexFractional) <-chan Curver {
	panic("not implemented yet")
	// http://kaj.uniwersytetradom.pl/prace/Biarcs.pdf
	// https://en.wikipedia.org/wiki/Arc_length
	// https://en.wikipedia.org/wiki/Ramer%E2%80%93Douglas%E2%80%93Peucker_algorithm
	// https://www.redblobgames.com/articles/curved-paths/
	// http://www.ryanjuckett.com/programming/biarc-interpolation/
	// https://stag-ws.zcu.cz/ws/services/rest/kvalifikacniprace/downloadPraceContent?adipIdno=17817
	// https://www.ajdesigner.com/phpcircle/circle_arc_length_s.php

	// Ok, so a few things to note:
	// The distance between center points on two adjacent hexes is 1.
	// Their shared edge has length 0.57735, or 1/(sqrt(3)).
	// A radius of 0.5 would make the arc blow out the top of the two
	// allowed hexes.

}

// biarc returns a list of circular arcs that connect pi to pe,
// with ti being the tangent at pi and te being the tangent at pe.
// This algorithm was adapted from "The use of Piecewise Circular Curves in Geometric
// Modeling" by Ulugbek Khudayarov.
func biarc(pi, ti, pe, te pos.HexFractional) (arcs []CircularArc) {
	// Tangents should be unit vectors.
	ti = ti.Normalize()
	te = te.Normalize()

	t := ti.Add(te)

	v := pe.Subtract(pi)

	// j is the joint point between the two arcs.
	var j pos.HexFractional
	// tj is the tangent at point j
	var tj pos.HexFractional
	// d is an intermediate discriminant value.
	var d float64
	// a is some intermediate constant.
	var a float64

	// This is the line segment case.
	if closeEnough(v.Normalize().DotProduct(t), 1.0) {
		return []CircularArc{
			CircularArc{pi, ti, pe},
		}
	}

	// Start and end tangents are parallel. (N2)
	if closeEnough(t.Length(), 2.0) {
		// Semicircle case. (N3)
		if closeEnough(v.DotProduct(ti), 0.0) {
			j = pi.Add(v.Multiply(0.5))
			tj = ti.Multiply(-1.0)

			return []CircularArc{
				CircularArc{pi, ti, j},
				CircularArc{j, tj, pe},
			}
		}
		// Calculate d from 2.6
		vdt := v.DotProduct(t)
		vl := v.Length()
		tl := t.Length()
		d = vdt*vdt + vl*vl*(4-tl*tl)
		// Calculate a from 2.7
		a = (math.Sqrt(d) - vdt) / (4 - tl*tl)
	} else { // N4
		// Calculate d from 2.6
		vdt := v.DotProduct(t)
		vl := v.Length()
		tl := t.Length()
		d = vdt*vdt + vl*vl*(4-tl*tl)
		// Calculate a from 2.8
		a = (vl * vl) / (4 * v.DotProduct(ti))
	}

	j = pi.Add(ti.Subtract(te).Multiply(a).Add(v).Multiply(0.5))
	tj = v.Subtract(t.Multiply(a)).Multiply(-2 * a)

	return []CircularArc{
		CircularArc{pi, ti, j},
		CircularArc{j, tj, pe},
	}
}
