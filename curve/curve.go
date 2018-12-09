package curve

import (
	"math"

	"github.com/erinpentecost/hexcoord/pos"
)

// SpinDirection of curve spin
type SpinDirection int

const (
	// CounterClockwise direction
	CounterClockwise SpinDirection = 1
	// Clockwise direction
	Clockwise SpinDirection = -1
	// NoSpin direction
	NoSpin SpinDirection = 0
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

	// Spin whether the curve is  in the clockwise (false)
	// or counterclockwise (true) direction. For lines, this will
	// return no spin
	Spin() SpinDirection
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

// Spin returns an error.
func (ls Line) Spin() SpinDirection {
	return NoSpin
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
	spin            SpinDirection
	cX              float64
	cY              float64
	piX             float64
	piY             float64
	piA             float64
	peX             float64
	peY             float64
	peA             float64
}

// lerpAngle traces an arc.
// spin is clockwise (false) or counterclockwise (true)
func lerpAngle(spin SpinDirection, a, b, t float64) float64 {
	if spin == CounterClockwise {
		return a + t*normalizeAngle(b-a)
	}
	return b + (1.0-t)*normalizeAngle(a-b)
}

// normalizeAngle places the angle in the range of pi to -pi.
func normalizeAngle(a float64) float64 {
	return a - 2*math.Pi*math.Floor((a+math.Pi)/(2*math.Pi))
}

// Sample returns a point on the curve.
// t is valid for 0 to 1, inclusive.
func (ac Arc) Sample(t float64) (position, tangent, curvature pos.HexFractional) {

	angle := lerpAngle(ac.spin, ac.piA, ac.peA, t)

	// sweep by some ratio of the maximal central angle to get position.
	// ptX := ac.cX + ac.radius*math.Cos(angle)
	// ptY := ac.cY + ac.radius*math.Sin(angle)
	unitPosition := pos.HexFractionalFromCartesian(math.Cos(angle), math.Sin(angle)).Normalize()
	position = unitPosition.Multiply(ac.radius).Add(ac.Center)

	// and tangent...
	if ac.spin == CounterClockwise {
		tangent = unitPosition.Rotate(pos.OriginFractional(), math.Pi/(2.0))
	} else {
		tangent = unitPosition.Rotate(pos.OriginFractional(), math.Pi/(-2.0))
	}

	// curvature points toward the center of the circle
	curvature = ac.Center.Subtract(position).Normalize().Multiply(ac.scalarCurvature)
	return
}

// Length returns the length of the curve.
func (ac Arc) Length() float64 {
	return ac.length
}

// Spin whether the curve is  in the clockwise (false)
// or counterclockwise (true) direction. For lines, this will
// return an error.
func (ac Arc) Spin() SpinDirection {
	return ac.spin
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

// angleDistance determines the difference between two angles in radians.
// spin is clockwise (false) or counterclockwise (true)
func angleDistance(start, end float64, spin SpinDirection) float64 {
	if spin == Clockwise {
		return angleDistance(end, start, CounterClockwise)
	}
	// at this point, only consider counterclockwise spin
	if start < end {
		return end - start
	}
	return end + (2*math.Pi - start)
}

func getSlope(p pos.HexFractional) float64 {
	pX, pY := p.ToCartesian()
	return -1.0 * pX / pY
}

// newArc creates a circular arc segment curve.
func newArc(pi, tiu, pe pos.HexFractional) Arc {
	// https://math.stackexchange.com/questions/996582/finding-circle-with-two-points-on-it-and-a-tangent-from-one-of-the-points

	// First line segment
	piX, piY := pi.ToCartesian()
	tiuOrthogonalSlope := getSlope(tiu)

	// Second line segment
	midX, midY := pos.LerpHexFractional(pi, pe, 0.5).ToCartesian()
	chordOrthogonalSlope := getSlope(pi.Subtract(pe))

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
	spin := Clockwise
	if area(pi, pi.Add(tiu), pe) < 0 {
		spin = CounterClockwise
	}

	// piA and peA are in the range 0 to 2pi
	centralAngle := angleDistance(piA, peA, spin)

	if centralAngle < 0.0 {
		panic("oh no")
	}

	return Arc{
		ca:              CircularArc{pi, tiu, pe},
		Center:          center,
		scalarCurvature: float64(1.0) / r,
		CentralAngle:    centralAngle,
		length:          r * centralAngle,
		radius:          r,
		spin:            spin,
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

// Spin returns the spin of the curve.
func (cs Piecewise) Spin() SpinDirection {
	if len(cs.segments) == 0 {
		return NoSpin
	}
	prev := cs.segments[0].Spin()
	if prev == NoSpin {
		return NoSpin
	}

	for _, s := range cs.segments {
		cSpin := s.Spin()
		if cSpin == NoSpin {
			return NoSpin
		}
		if cSpin != prev {
			return NoSpin
		}
	}
	return prev
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
