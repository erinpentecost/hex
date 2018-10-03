package hexcoord

import (
	"math"
)

// Curver is a continuous curve segment.
// It can be used to draw a curve.
type Curver interface {
	// Sample returns a point on the curve.
	// t is valid for 0 to 1, inclusive.
	Sample(t float64) (position, tangent, curvature HexFractional)

	// Length returns the length of the curve.
	Length() float64
}

// lineCurve is a Curver.
type lineCurve struct {
	i      HexFractional
	e      HexFractional
	length float64
	slope  HexFractional
}

// Sample returns a point on the curve.
// t is valid for 0 to 1, inclusive.
func (ls lineCurve) Sample(t float64) (position, tangent, curvature HexFractional) {
	position = LerpHexFractional(ls.i, ls.e, t)
	tangent = ls.slope
	curvature = HexFractional{0.0, 0.0}
	return
}

// Length returns the length of the curve.
func (ls lineCurve) Length() float64 {
	return ls.i.DistanceTo(ls.e)
}

// newLine creates a line segment curve.
// Inputs are start and end points.
func newLine(i, e HexFractional) lineCurve {

	return lineCurve{
		i:      i,
		e:      e,
		length: i.DistanceTo(e),
		slope:  e.Subtract(i).Normalize(),
	}
}

// arcCurve is a Curver.
type arcCurve struct {
	ca              CircularArc
	center          HexFractional
	scalarCurvature float64
	centralAngle    float64
	length          float64
	longWay         bool
	direction       float64
}

// Sample returns a point on the curve.
// t is valid for 0 to 1, inclusive.
func (ac arcCurve) Sample(t float64) (position, tangent, curvature HexFractional) {

	// sweep by some ratio of the maximal central angle to get position.
	position = ac.ca.I.Rotate(ac.center, ac.direction*t*ac.centralAngle)

	origin := HexOrigin().ToHexFractional()
	tangent = position.Subtract(ac.center).Rotate(origin, ac.direction*math.Pi/2).Normalize()

	// curvature points toward the center of the circle
	curvature = position.Subtract(ac.center).Normalize().Multiply(ac.scalarCurvature)
	return
}

// Length returns the length of the curve.
func (ac arcCurve) Length() float64 {
	return ac.length
}

// area determines the triangular area between three points.
// It's not what you'd expect (euclidean). This is just here
// to aid in testing for collinearity and clockwise/cc detection.
// http://mathworld.wolfram.com/Collinear.html
func area(a, b, c HexFractional) float64 {
	return a.Q*(b.R-c.R) + b.Q*(c.R-a.R) + c.Q*(a.R-b.R)
}

// intersection gets the intersecting point described by two
// lines. This is all in cartersian coordinates.
func intersection(ax, ay, am, bx, by, bm float64) (ix, iy float64) {
	if math.IsNaN(am) || math.IsInf(am, 0) {
		ix = ax
		iy = bm*(ix-bx) + by
		return
	} else if math.IsNaN(bm) || math.IsInf(bm, 0) {
		ix = bx
		iy = am*(ix-ax) + ay
		return
	}

	ix = (am*ax - bm*bx - ay + by) / (am - bm)
	iy = am*ix - am*bx + by
	return
}

// newArc creates a circular arc segment curve.
func newArc(pi, tiu, pe HexFractional) arcCurve {
	// https://math.stackexchange.com/questions/996582/finding-circle-with-two-points-on-it-and-a-tangent-from-one-of-the-points
	piX, piY := pi.ToCartesian()

	midX, midY := LerpHexFractional(pi, pe, 0.5).ToCartesian()

	tiuX, tiuY := tiu.ToCartesian()
	tiuOrthogonalSlope := -1.0 * tiuX / tiuY

	chordX, chordY := pe.Subtract(pi).ToCartesian()
	chordOrthogonalSlope := -1.0 * chordX / chordY

	// Find the intersection of two lines:
	// pi with slope tanOrth
	// mid with slope chordOrth
	// This gets the circle center point.
	centerX, centerY := intersection(piX, piY, tiuOrthogonalSlope, midX, midY, chordOrthogonalSlope)

	// Find the center by projecting the midpoint on
	// the chord to a vector orthogonal to the tangent.
	center := HexFractionalFromCartesian(centerX, centerY)

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
	// I think this is the "scalar triple product"
	var direction float64
	clockwise := math.Signbit(area(pi, center, pe))
	if longWay {
		clockwise = !clockwise
	}
	if clockwise {
		direction = -1.0
	} else {
		direction = 1.0
	}

	return arcCurve{
		ca:              CircularArc{pi, tiu, pe},
		center:          center,
		scalarCurvature: float64(1.0) / radius.Length(),
		centralAngle:    centralAngle,
		length:          radius.Length() * centralAngle,
		longWay:         longWay,
		direction:       direction,
	}
}

// combinationCurve is a Curver.
type combinationCurve struct {
	segments []Curver
	length   float64
}

// Sample returns a point on the curve.
// t is valid for 0 to 1, inclusive.
func (cs combinationCurve) Sample(t float64) (position, tangent, curvature HexFractional) {
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
func (cs combinationCurve) Length() float64 {
	return cs.length
}

// JoinCurves creates a multipart curve.
func JoinCurves(curves ...Curver) Curver {
	// Don't wrap a single element.
	if len(curves) == 1 {
		return curves[0]
	}

	cs := combinationCurve{
		segments: curves,
		length:   float64(0.0),
	}

	for _, a := range curves {
		cs.length += a.Length()
	}

	return cs
}

// Curve converts a circular arc into a sample-able curve.
func (ca CircularArc) Curve() Curver {
	if closeEnough(area(ca.I, ca.T.Add(ca.I), ca.E), 0.0) {
		// This is the line segment case, where ca.i + ca.tiu is collinear with ca.e.
		return newLine(ca.I, ca.E)
	}
	// This is the circular arc case.
	return newArc(ca.I, ca.T, ca.E)
}
