package hexcoord

import (
	"math"
)

// circularArc defines a circular arc in three vectors.
type circularArc struct {
	// i is the initial point.
	i HexFractional
	// tiu is the tangent unit vector at the initial point.
	tiu HexFractional
	// e is the end point.
	e HexFractional
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
func SmoothPath(done <-chan interface{}, ti, te, path []HexFractional) <-chan CurveSegmenter {
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
func biarc(pi, ti, pe, te HexFractional) (arcs []circularArc) {
	// Tangents should be unit vectors.
	ti = ti.Normalize()
	te = te.Normalize()

	t := ti.Add(te)

	v := pe.Subtract(pi)

	// j is the joint point between the two arcs.
	var j HexFractional
	// tj is the tangent at point j
	var tj HexFractional
	// d is an intermediate discriminant value.
	var d float64
	// a is some intermediate constant.
	var a float64

	// This is the line segment case.
	if closeEnough(v.Normalize().DotProduct(t), 1.0) {
		return []circularArc{
			circularArc{pi, ti, pe},
		}
	}

	// Start and end tangents are parallel. (N2)
	if closeEnough(t.Length(), 2.0) {
		// Semicircle case. (N3)
		if closeEnough(v.DotProduct(ti), 0.0) {
			j = pi.Add(v.Multiply(0.5))
			tj = ti.Multiply(-1.0)

			return []circularArc{
				circularArc{pi, ti, j},
				circularArc{j, tj, pe},
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

	return []circularArc{
		circularArc{pi, ti, j},
		circularArc{j, tj, pe},
	}
}
