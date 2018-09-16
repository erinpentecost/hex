package hexcoord

import (
	"math"
)

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
func SmoothPath(ti, te, path []HexFractional) func(t float64) (m0, m1, m2 HexFractional) {
	panic("not implemented yet")
	// http://kaj.uniwersytetradom.pl/prace/Biarcs.pdf
	// https://en.wikipedia.org/wiki/Arc_length
	// https://en.wikipedia.org/wiki/Ramer%E2%80%93Douglas%E2%80%93Peucker_algorithm
	// https://www.redblobgames.com/articles/curved-paths/
	// http://www.ryanjuckett.com/programming/biarc-interpolation/
	// https://stag-ws.zcu.cz/ws/services/rest/kvalifikacniprace/downloadPraceContent?adipIdno=17817

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
func GenerateBiarc(pi, ti, pe, te HexFractional, r float64) func(t float64) (m0, m1, m2 HexFractional) {
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
func combineSegments(segmentFns [](func(t float64) (m0, m1, m2 HexFractional))) func(t float64) (m0, m1, m2 HexFractional) {
	return func(t float64) (m0, m1, m2 HexFractional) {
		// segment is the segment we are in.
		segment := int(t / float64(len(segmentFns)))
		remappedT := t / float64(segment+1)
		return segmentFns[segment](remappedT)
	}
}

// lineSegment creates a lineSegment function.
// Inputs are start and end points.
// Outputs are point, tangent, and curvature.
func lineSegment(pi, pe HexFractional) func(t float64) (m0, m1, m2 HexFractional) {
	slope := pi.Subtract(pe).Normalize()
	return func(t float64) (m0, m1, m2 HexFractional) {
		m0 = LerpHexFractional(pi, pe, t)
		m1 = slope
		m2 = HexFractional{0.0, 0.0}
		return
	}
}

// arc creates an arc function.
// Inputs are start point, curvature point, and end point.
// The curvature point is the intersection of the tangents of the start and end points.
// Outputs are point, tangent, and curvature.
func arc(pi, pr, pe HexFractional) func(t float64) (m0, m1, m2 HexFractional) {
	panic("not yet implemented")
}
