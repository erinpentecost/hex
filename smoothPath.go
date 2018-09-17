package hexcoord

import (
	"math"
)

// CurveSegmenter is a continuous curve segment.
type CurveSegmenter interface {
	// Sample returns a point on the curve.
	// t is valid for 0 to 1, inclusive.
	Sample(t float64) (position, tangent, curvature HexFractional)

	// Length returns the length of the curve.
	Length() float64
}

type curveSegmenterImpl struct {
	sample func(t float64) (position, tangent, curvature HexFractional)
	length float64
}

func (csi curveSegmenterImpl) Sample(t float64) (position, tangent, curvature HexFractional) {
	return csi.sample(t)
}

func (csi curveSegmenterImpl) Length() float64 {
	return csi.length
}

type circularArc struct {
	// i is the initial point.
	i HexFractional
	// tiu is the tangent unit vector at the initial point.
	tiu HexFractional
	// e is the end point.
	e HexFractional
}

func (ca circularArc) CurveSegmenter() CurveSegmenter {
	// This is split into 3 cases in an attempt to work
	// around inaccuracies introduced by using floating points.
	v := ca.e.Subtract(ca.i)
	vtDot := v.Normalize().DotProduct(ca.tiu)
	if closeEnough(vtDot, 1.0) {
		// This is the line segment case, where ca.i + ca.tiu is collinear with ca.e.
		return lineSegment(ca.i, ca.e)
	} else {
		// This is the circular arc case.
		panic("not implemented yet")
	}
}

// lineSegment creates a lineSegment curve.
// Inputs are start and end points.
// Outputs are point, tangent, and curvature.
func lineSegment(pi, pe HexFractional) CurveSegmenter {
	slope := pi.Subtract(pe).Normalize()

	return curveSegmenterImpl{
		sample: func(t float64) (position, tangent, curvature HexFractional) {
			position = LerpHexFractional(pi, pe, t)
			tangent = slope
			curvature = HexFractional{0.0, 0.0}
			return
		},
		length: pi.DistanceTo(pe),
	}
}

// this should be dropped and generalized
func semiCircleSegment(pi, tiu, pe HexFractional) CurveSegmenter {
	diameter := pi.DistanceTo(pe)
	center := pe.Subtract(pi).Multiply(0.5).Add(pi)
	arcLength := math.Pi * diameter / 2.0
	scalarCurvature := 2.0 / diameter
	centralAngle := math.Pi

	// t = 0 is arcLength = 0 = arclength to pi
	// t = 1 is arcLength = max arcLength = arclength to pe

	return curveSegmenterImpl{
		sample: func(t float64) (position, tangent, curvature HexFractional) {
			// sweep by some ratio of the maximal central angle to get position.
			position = pi.Rotate(center, t*centralAngle)

			tangent = getTangent(center, position)

			// curvature points toward the center of the circle
			curvature = position.Subtract(center).Normalize().Multiply(scalarCurvature * (-1.0))
			return
		},
		length: arcLength,
	}
}

// getTangent returns the tangent at arcPoint for the circle
// defined by the given center and a radius equal to
// arcPoint.Subtract(center).Length().
func getTangent(center, arcPoint HexFractional) HexFractional {
	panic("not implemented yet")
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
func SmoothPath(ti, te, path []HexFractional) CurveSegmenter {
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

// CombineSegments presents a slice of segments as a single segment.
func CombineSegments(arcs []CurveSegmenter) CurveSegmenter {

	totalLength := float64(0.0)
	for _, a := range arcs {
		totalLength += a.Length()
	}

	return curveSegmenterImpl{
		sample: func(t float64) (position, tangent, curvature HexFractional) {
			lenT := t * totalLength
			// determine which sub-segment t lands us in.
			prevLength := float64(0.0)
			runningLength := float64(0.0)
			for _, a := range arcs {
				runningLength += a.Length()
				if lenT <= runningLength {
					// we are in the current segment
					// now we need to remap t for it

					remappedT := (lenT - prevLength) / runningLength
					return a.Sample(remappedT)
				}
				prevLength += a.Length()
			}
			panic("t is out of scope")
		},
		length: totalLength,
	}
}
