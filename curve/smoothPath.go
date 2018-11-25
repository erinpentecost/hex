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
// As output, it will return a piecewise collection of circular
// arcs that connect those hexes with G1 continuity.
// These arcs can be converted to parameterized curves with
// the Curve() function.
func SmoothPath(ti pos.HexFractional, te pos.HexFractional, path []pos.HexFractional) []CircularArc {

	// http://kaj.uniwersytetradom.pl/prace/Biarcs.pdf
	// https://en.wikipedia.org/wiki/Arc_length
	// https://en.wikipedia.org/wiki/Ramer%E2%80%93Douglas%E2%80%93Peucker_algorithm
	// https://www.redblobgames.com/articles/curved-paths/
	// http://www.ryanjuckett.com/programming/biarc-interpolation/
	// https://stag-ws.zcu.cz/ws/services/rest/kvalifikacniprace/downloadPraceContent?adipIdno=17817
	// https://www.ajdesigner.com/phpcircle/circle_arc_length_s.php

	// If there are 1 or fewer points, we are already
	// at the target path.
	if len(path) < 2 {
		return make([]CircularArc, 0, 0)
	}

	curves := make([]CircularArc, 0, 2*len(path))

	// Find tangents for each position.
	tangents := make([]pos.HexFractional, len(path), len(path))
	tangents[0] = ti
	tangents[len(path)-1] = te
	for p := 1; p < len(path)-1; p++ {
		tangents[p] = approximateTangent(path[p-1], path[p], path[p+1])
	}

	// Generate biarcs for each pair of points.
	for i := 0; i < len(path)-1; i++ {
		for _, b := range Biarc(path[i], tangents[i], path[i+1], tangents[i+1]) {
			curves = append(curves, b)
		}
	}

	return curves
}

// This algorithm was adapted from "The use of Piecewise Circular Curves in Geometric
// Modeling" by Ulugbek Khudayarov.
func approximateTangent(p0, p1, p2 pos.HexFractional) pos.HexFractional {
	a := p1.Subtract(p0)
	b := p2.Subtract(p1)
	aLen := a.Length()
	bLen := b.Length()

	return a.Multiply(bLen / aLen).Add(b.Multiply(aLen / bLen))
}

// Biarc returns a list of circular arcs that connect pi to pe,
// with ti being the tangent at pi and te being the tangent at pe.
// This algorithm was adapted from "The use of Piecewise Circular Curves in Geometric
// Modeling" by Ulugbek Khudayarov.
func Biarc(pi, ti, pe, te pos.HexFractional) (arcs []CircularArc) {
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
	// dumb version
	//_, tj, _ := CircularArc{pi, ti, j}.Curve().Sample(1.0)

	return []CircularArc{
		CircularArc{pi, ti, j},
		CircularArc{j, tj, pe},
	}
}
