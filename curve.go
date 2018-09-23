package hexcoord

import (
	"math"
)

// CurveSegmenter is a continuous curve segment.
// It can be used to draw a curve.
type CurveSegmenter interface {
	// Sample returns a point on the curve.
	// t is valid for 0 to 1, inclusive.
	Sample(t float64) (position, tangent, curvature HexFractional)

	// Length returns the length of the curve.
	Length() float64
}

// LineSegment is a CurveSegmenter.
type LineSegment struct {
	i      HexFractional
	e      HexFractional
	length float64
	slope  HexFractional
}

// Sample returns a point on the curve.
// t is valid for 0 to 1, inclusive.
func (ls LineSegment) Sample(t float64) (position, tangent, curvature HexFractional) {
	position = LerpHexFractional(ls.i, ls.e, t)
	tangent = ls.slope
	curvature = HexFractional{0.0, 0.0}
	return
}

// Length returns the length of the curve.
func (ls LineSegment) Length() float64 {
	return ls.i.DistanceTo(ls.e)
}

// NewLineSegment creates a line segment curve.
// Inputs are start and end points.
func NewLineSegment(i, e HexFractional) LineSegment {

	return LineSegment{
		i:      i,
		e:      e,
		length: i.DistanceTo(e),
		slope:  i.Subtract(e).Normalize(),
	}
}

// ArcSegment is a CurveSegmenter.
type ArcSegment struct {
	ca              circularArc
	center          HexFractional
	scalarCurvature float64
	centralAngle    float64
	length          float64
	longWay         bool
	direction       float64
}

// Sample returns a point on the curve.
// t is valid for 0 to 1, inclusive.
func (ac ArcSegment) Sample(t float64) (position, tangent, curvature HexFractional) {
	// sweep by some ratio of the maximal central angle to get position.
	position = ac.ca.i.Rotate(ac.center, t*ac.centralAngle).Multiply(ac.direction)

	// This should be perpendicular to the radius,
	// but the direction may be wrong.
	tangentLine := HexFractional{
		Q: -1 * position.R,
		R: position.Q,
	}

	// By projecting the end position onto the tangent line,
	// I get a tangent vector that is pointing toward it.
	if tangentLine.DotProduct(ac.ca.e) != 0 {
		tangent = ac.ca.e.ProjectOn(tangentLine).Normalize()
	} else {
		// If we are on the end line, we need to use start position
		// and then reverse.
		tangent = ac.ca.i.ProjectOn(tangentLine).Multiply(-1).Normalize()
	}

	// curvature points toward the center of the circle
	curvature = position.Subtract(ac.center).Normalize().Multiply(ac.scalarCurvature * (-1.0))
	return
}

// Length returns the length of the curve.
func (ac ArcSegment) Length() float64 {
	return ac.length
}

// NewArcSegment creates a circular arc segment curve.
func NewArcSegment(pi, tiu, pe HexFractional) ArcSegment {

	// Find the center by projecting the midpoint on
	// the chord to a vector orthogonal to the tangent.
	center := LerpHexFractional(pi, pe, 0.5).ProjectOn(HexFractional{
		Q: -1 * tiu.R,
		R: tiu.Q,
	}.Add(pi))

	radius := pi.Subtract(center)

	// This gets the internal angle 100% of the time.
	centralAngle := radius.AngleTo(pe.Subtract(center))
	longWay := false
	// But I may need the complimentary angle instead.
	if pi.Add(tiu).Subtract(center).AngleTo(pe.Subtract(center)) > centralAngle {
		centralAngle = 2*math.Pi - centralAngle
		longWay = true
	}

	// Determine clockwise vs counterclockwise
	// For a small central angle, the sign of the area for the tiangle works.
	var direction float64
	clockwise := math.Signbit((pi.Q-center.Q)*(pe.R-center.R) - (pi.R-center.R)*(pe.Q-center.Q))
	if longWay {
		clockwise = !clockwise
	}
	if clockwise {
		direction = -1.0
	} else {
		direction = 1.0
	}

	return ArcSegment{
		ca:              circularArc{pi, tiu, pe},
		center:          center,
		scalarCurvature: float64(1.0) / radius.Length(),
		centralAngle:    centralAngle,
		length:          radius.Length() * centralAngle,
		longWay:         longWay,
		direction:       direction,
	}
}

// combinationSegment is a CurveSegmenter.
type combinationSegment struct {
	segments []CurveSegmenter
	length   float64
}

// Sample returns a point on the curve.
// t is valid for 0 to 1, inclusive.
func (cs combinationSegment) Sample(t float64) (position, tangent, curvature HexFractional) {
	lenT := t * cs.length
	// determine which sub-segment t lands us in.
	prevLength := float64(0.0)
	runningLength := float64(0.0)
	for _, a := range cs.segments {
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
}

// Length returns the length of the curve.
func (cs combinationSegment) Length() float64 {
	return cs.length
}

// JoinSegments creates a multipart curve.
func JoinSegments(arcs ...CurveSegmenter) CurveSegmenter {

	// Don't wrap a single element.
	if len(arcs) == 1 {
		return arcs[0]
	}

	cs := combinationSegment{
		segments: arcs,
		length:   float64(0.0),
	}

	for _, a := range arcs {
		cs.length += a.Length()
	}

	return cs
}

// newCurveSegmenter converts a circular arc into a sample-able curve.
func newCurveSegmenter(ca circularArc) CurveSegmenter {
	v := ca.e.Subtract(ca.i)
	vtDot := v.Normalize().DotProduct(ca.tiu)
	if closeEnough(vtDot, 1.0) {
		// This is the line segment case, where ca.i + ca.tiu is collinear with ca.e.
		return NewLineSegment(ca.i, ca.e)
	}
	// This is the circular arc case.
	return NewArcSegment(ca.i, ca.tiu, ca.e)
}
