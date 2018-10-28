package curve

import (
	"math"

	"github.com/erinpentecost/hexcoord/pos"
)

// Curver is a continuous curve segment.
// It can be used to draw a curve.
type Curver interface {
	// Sample returns a point on the curve.
	// t is valid for 0 to 1, inclusive.
	// curvature points toward the "center" of the curve.
	Sample(t float64) (position, tangent, curvature pos.HexFractional)

	// Length returns the length of the curve.
	Length() float64
}

// Line is a Curver.
type Line struct {
	i      pos.HexFractional
	e      pos.HexFractional
	length float64
	slope  pos.HexFractional
}

// Sample returns a point on the curve.
// t is valid for 0 to 1, inclusive.
func (ls Line) Sample(t float64) (position, tangent, curvature pos.HexFractional) {
	position = pos.LerpHexFractional(ls.i, ls.e, t)
	tangent = ls.slope
	curvature = pos.OriginFractional()
	return
}

// Length returns the length of the curve.
func (ls Line) Length() float64 {
	return ls.i.DistanceTo(ls.e)
}

// newLine creates a line segment curve.
// Inputs are start and end points.
func newLine(i, e pos.HexFractional) Line {

	return Line{
		i:      i,
		e:      e,
		length: i.DistanceTo(e),
		slope:  e.Subtract(i).Normalize(),
	}
}

// Arc is a Curver.
type Arc struct {
	ca              CircularArc
	Center          pos.HexFractional
	scalarCurvature float64
	CentralAngle    float64
	length          float64
	radius          float64
	Spin            bool
	cX              float64
	cY              float64
	piX             float64
	piY             float64
	piA             float64
	peX             float64
	peY             float64
	peA             float64
}

func lerpAngle(a, b, t float64) float64 {
	return a + t*normalizeAngle(b-a)
}

// normalizeAngle places the angle in the range of pi to -pi.
func normalizeAngle(a float64) float64 {
	return a - 2*math.Pi*math.Floor((a+math.Pi)/(2*math.Pi))
}

// Sample returns a point on the curve.
// t is valid for 0 to 1, inclusive.
func (ac Arc) Sample(t float64) (position, tangent, curvature pos.HexFractional) {

	angle := lerpAngle(ac.piA, ac.peA, t)

	// sweep by some ratio of the maximal central angle to get position.
	// ptX := ac.cX + ac.radius*math.Cos(angle)
	// ptY := ac.cY + ac.radius*math.Sin(angle)
	unitPosition := pos.HexFractionalFromCartesian(math.Cos(angle), math.Sin(angle)).Normalize()
	position = unitPosition.Multiply(ac.radius).Add(ac.Center)

	// and tangent...
	// todo: add or subtract 90 degrees?
	if ac.Spin {
		tangent = unitPosition.Rotate(pos.OriginFractional(), math.Pi/(-2.0))
	} else {
		tangent = unitPosition.Rotate(pos.OriginFractional(), math.Pi/2.0)
	}

	// curvature points toward the center of the circle
	curvature = ac.Center.Subtract(position).Normalize().Multiply(ac.scalarCurvature)
	return
}

// Length returns the length of the curve.
func (ac Arc) Length() float64 {
	return ac.length
}

// area determines the triangular area between three points.
// It's not what you'd expect (euclidean). This is just here
// to aid in testing for collinearity and clockwise/cc detection.
// http://mathworld.wolfram.com/Collinear.html
func area(a, b, c pos.HexFractional) float64 {
	return a.Q*(b.R-c.R) + b.Q*(c.R-a.R) + c.Q*(a.R-b.R)
}

// intersection gets the intersecting point described by two
// lines. This is all in cartersian coordinates.
func intersection(ax, ay, am, bx, by, bm float64) (ix, iy float64) {
	// y = am (x − ax) + ay
	// x = (ax*am - ay + y)/am and m!=0
	// x,y are the coordinates of any point on the line
	// am is the slope of the line
	// ax, ay are the x and y coordinates of the given point P that defines the line

	if math.IsNaN(am) || math.IsInf(am, 0) {
		ix = ax
		iy = bm*(ix-bx) + by
		return
	} else if math.IsNaN(bm) || math.IsInf(bm, 0) {
		ix = bx
		iy = am*(ix-ax) + ay
		return
	}

	// am (x − ax) + ay = bm (x − bx) + by
	// a (x − b) + c = m (x − n) + o
	// x = (a b - c - m n + o)/(a - m) and a!=m
	// x = (am*ax - ay - bm*bx + by)/(am - bm) and a!=m
	// y = am (x − ax) + ay
	// y = a (x − b) + c

	ix = (am*ax - ay - bm*bx + by) / (am - bm)
	iy = am*ix - am*ax + ay
	return
}

// getAngle returns the angle to the x axis for a cartesian vector.
func getAngle(x, y float64) float64 {
	switch getQuadrant(x, y) {
	case 1: // This is the cah rule, and is only valid for acute angles.
		denom := math.Sqrt(math.Pow(x, 2.0) + math.Pow(y, 2.0))
		return math.Acos((x) / denom)
	case 2:
		return math.Pi - getAngle(-1.0*x, y)
	case 3:
		return math.Pi + getAngle(-1.0*x, -1.0*y)
	case 4:
		return 2*math.Pi - getAngle(x, -1.0*y)
	default:
		panic("There are only 4 quadrants")
	}
}

// getQuadrant returns the quadrant that the cartesian vector represented by
// x and y is in. Returns a value between 1 and 4, inclusive.
//   4 | 1
//   3 | 2
func getQuadrant(x, y float64) int {
	xPos := !math.Signbit(x)
	yPos := !math.Signbit(y)
	if xPos && yPos {
		return 1
	} else if yPos {
		return 2
	} else if xPos {
		return 3
	} else {
		return 4
	}
}

// getSpin determines if an arc defined by some point (px, py) and the
// tangent (tx, ty) is going in the clockwise (false) or counterclockwise (true)
// direction. Keep in mind that the bottom of your monitor is in the positive y
// direction.
func getSpin(py, px, ty, tx float64) bool {
	if tx != 0 {
		if py >= 0 {
			return tx < 0
		}
		return tx > 0
	} else if ty != 0 {
		if px >= 0 {
			return ty > 0
		}
		return ty < 0
	}
	panic("Can't determine direction")
}

// newArc creates a circular arc segment curve.
func newArc(pi, tiu, pe pos.HexFractional) Arc {
	// https://math.stackexchange.com/questions/996582/finding-circle-with-two-points-on-it-and-a-tangent-from-one-of-the-points

	// First line segment
	piX, piY := pi.ToCartesian()
	tiuX, tiuY := tiu.ToCartesian()
	tiuOrthogonalSlope := -1.0 * tiuX / tiuY

	// Second line segment
	midX, midY := pos.LerpHexFractional(pi, pe, 0.5).ToCartesian()
	chordX, chordY := pi.Subtract(pe).ToCartesian()
	chordOrthogonalSlope := -1.0 * chordX / chordY

	// Find the intersection of two lines:
	// pi with slope tanOrth
	// mid with slope chordOrth
	// This gets the circle center point.
	// Example:
	//   -2.414 = (x-1.478) / (y+0.612)
	//   -1 = (x-1.224) / (y+1.224)
	//    intersection should be 0,0
	centerX, centerY := intersection(piX, piY, tiuOrthogonalSlope, midX, midY, chordOrthogonalSlope)
	center := pos.HexFractionalFromCartesian(centerX, centerY)

	radius := pi.Subtract(center)
	r := radius.Length()

	// Get start and stop angles.
	// https://math.stackexchange.com/questions/1144159/parametric-equation-of-an-arc-with-given-radius-and-two-points
	piA := getAngle(piX-centerX, piY-centerY)
	peX, peY := pe.ToCartesian()
	peA := getAngle(peX-centerX, peY-centerY)

	// Determine spin direction.
	// clockwise (false) or counterclockwise (true)
	spin := getSpin(piY, piX, tiuY, tiuX)

	// piA and peA are in the range 0 to 2pi
	var centralAngle float64
	if spin {
		centralAngle = peA - piA
	} else {
		centralAngle = piA - peA
	}

	if centralAngle < 0.0 {
		centralAngle = centralAngle + 2*math.Pi
	}

	return Arc{
		ca:              CircularArc{pi, tiu, pe},
		Center:          center,
		scalarCurvature: float64(1.0) / r,
		CentralAngle:    centralAngle,
		length:          r * centralAngle,
		radius:          r,
		Spin:            spin,
		cX:              centerX,
		cY:              centerY,
		piX:             piX,
		piY:             piY,
		piA:             piA,
		peX:             peX,
		peY:             peY,
		peA:             peA,
	}
}

// Piecewise is a Curver.
type Piecewise struct {
	segments []Curver
	length   float64
}

// Sample returns a point on the curve.
// t is valid for 0 to 1, inclusive.
func (cs Piecewise) Sample(t float64) (position, tangent, curvature pos.HexFractional) {
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
func (cs Piecewise) Length() float64 {
	return cs.length
}

// Join creates a multipart curve.
// No assertion is made that the input curves are
// connected.
func Join(curves ...Curver) Piecewise {

	// Store all segments.
	cs := Piecewise{
		segments: curves,
		length:   float64(0.0),
	}

	// Determine full length.
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

func closeEnough(a, b float64) bool {
	if a == b {
		return true
	}
	return math.Abs(a-b) < 1e-10
}