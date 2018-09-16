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
	if closeEnough(vtDot, 0.0) {
		// This is the semicircle case.
		return semiCircleSegment(ca.i, ca.tiu, ca.e)
	} else if closeEnough(vtDot, 1.0) {
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

			// TODO
			tangent = 0.0

			// curvature points toward the center of the circle
			curvature = position.Subtract(center).Normalize().Multiply(scalarCurvature * (-1.0))
			return
		},
		length: arcLength,
	}
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

// GenerateBiarc returns a function that represents a biarc.
// pi is the initial point. ti is the starting tangent of that point.
// pe is the end point. te is the ending tangent of that point.
// r is a ratio that narrows the result set to one function. Generally,
// a value of 1.0 is good enough, and tries to keep the curvatures of
// the two arcs close.
func GenerateBiarc(pi, ti, pe, te HexFractional, r float64) (arcs []circularArc) {
	// Tangents should be unit vectors.
	tiu := ti.Normalize()
	teu := te.Normalize()

	// r should be positive.
	if r <= 0 {
		panic("r must be positive")
	}

	// r = α/β, and all terms are positive.
	// I need to find α and β.

	// Build up quadratic formula for β.
	v := pi.Subtract(pe)
	a := v.DotProduct(v)
	b := 2 * v.DotProduct(tiu.Multiply(r).Add(teu))
	c := 2 * r * (tiu.DotProduct(teu) - 1)

	// Is this a special case?
	// If the vectors v, ti, te are collinear, this is a line segment.
	if c == 0 {
		panic("start and end tangents (ti and te) aren't allowed to face in the same direction")
		// this can be fixed by returning two semicircles in an s shape.
	} else if b == 0 {
		//panic("this needs 4 arcs")
		// this is probably the straight-line case.
		return lineSegment(pi, pe)
	}

	// Solve for β.
	// Pick a positive root, if able.
	beta := (math.Sqrt(b*b-4.0*a*c) - b) / (2.0 * a)
	if beta <= 0 {
		beta = -1.0 * (math.Sqrt(b*b-4.0*a*c) + b) / (2.0 * a)
	}

	if beta == 0 {
		panic("can't find a nonzero root")
	}

	// Solve for α.
	alpha := r * beta

	// Arc control points.
	// These need weights assigned to them?
	arcFirst := pi.Add(ti.Multiply(alpha))
	arcSecond := pe.Subtract(te.Multiply(beta))

	// Inflection / connection point.
	inflection := arcFirst.Multiply(beta / (alpha + beta)).Add(arcSecond.Multiply(alpha / (alpha + beta)))

	panic("not yet implemented")
}

// combineSegments presents a slice of segments as a single segment.
// Each segment in the slice is given equal weight.
// If these were travel paths, the same amount of time would be spent in
// each segment. Movement would be faster for shorter segments.
// TODO: Weight (re: t) to be based on path length instead.
func combineSegments(arcs []CurveSegmenter) CurveSegmenter {
	return func(t float64) (m0, m1, m2 HexFractional) {
		// segment is the segment we are in.
		segment := int(t / float64(len(segmentFns)))
		remappedT := t / float64(segment+1)
		return segmentFns[segment](remappedT)
	}
}
