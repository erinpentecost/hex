package hexcoord_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/erinpentecost/hexcoord"
	"github.com/stretchr/testify/assert"
)

func printArc(c hexcoord.CircularArc) string {
	return fmt.Sprintf("arc(I:%v,%vy; T:%v%v; E:%v%v)", c.I.Q, c.I.R, c.T.Q, c.T.R, c.E.Q, c.E.R)
}

func assertSample(t *testing.T, prefix interface{}, f float64, c hexcoord.Curver, sp, st, sc hexcoord.HexFractional) {

	cp, ct, cc := c.Sample(f)
	assert.True(t, sp.AlmostEquals(cp), fmt.Sprintf("%v: At sample %v, got position %v but expected %v.", prefix, f, cp.ToString(), sp.ToString()))
	assert.True(t, st.AlmostEquals(ct), fmt.Sprintf("%v: At sample %v, got tangent %v but expected %v.", prefix, f, ct.ToString(), st.ToString()))
	assert.True(t, sc.AlmostEquals(cc), fmt.Sprintf("%v: At sample %v, got curvature %v but expected %v.", prefix, f, cc.ToString(), sc.ToString()))
}

func TestLineCurve(t *testing.T) {

	done := make(chan interface{})
	defer close(done)

	testHexes := hexcoord.AreaToSlice(hexcoord.Origin().SpiralArea(done, 4))
	origin := hexcoord.Origin().ToHexFractional()

	for _, i := range testHexes {
		for _, e := range testHexes {
			if i == e {
				continue
			}

			tangent := e.ToHexFractional().Subtract(i.ToHexFractional()).Normalize()

			line := hexcoord.CircularArc{
				I: i.ToHexFractional(),
				T: tangent,
				E: e.ToHexFractional(),
			}

			curve := line.Curve()

			assertSample(t, line, 0.0, curve, line.I, line.T, origin)
			assertSample(t, line, 0.1, curve, hexcoord.LerpHexFractional(line.I, line.E, 0.1), line.T, origin)
			assertSample(t, line, 0.5, curve, hexcoord.LerpHexFractional(line.I, line.E, 0.5), line.T, origin)
			assertSample(t, line, 0.75, curve, hexcoord.LerpHexFractional(line.I, line.E, 0.75), line.T, origin)
			assertSample(t, line, 1.0, curve, line.E, line.T, origin)
		}
	}

}

func lerpFloat(a, b, t float64) float64 {
	return a*(1.0-t) + b*t
}

func TestUnitCircle(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	sampleStep := math.Pi / 6

	testHexes := hexcoord.AreaToSlice(hexcoord.Origin().RingArea(done, 1))
	for _, i := range testHexes {
		init := i.ToHexFractional().Normalize()
		for sweep := sampleStep; sweep < math.Pi*2; sweep = sweep + sampleStep {
			unitCircle(t, init, sweep, false)
		}
	}
}

func TestUnitCircleReversed(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	sampleStep := math.Pi / 6

	testHexes := hexcoord.AreaToSlice(hexcoord.Origin().RingArea(done, 1))
	for _, i := range testHexes {
		init := i.ToHexFractional().Normalize()
		for sweep := sampleStep; sweep < math.Pi*2; sweep = sweep + sampleStep {
			unitCircle(t, init, sweep, true)
		}
	}
}

func unitCircle(t *testing.T, fp hexcoord.HexFractional, sweepRadians float64, reverse bool) {
	origin := hexcoord.Origin().ToHexFractional()
	firstPoint := fp.Normalize()

	getTan := func(a hexcoord.HexFractional) hexcoord.HexFractional {
		return a.Rotate(origin, -1.0*math.Pi/2).Normalize()
	}
	if reverse {
		getTan = func(a hexcoord.HexFractional) hexcoord.HexFractional {
			return a.Rotate(origin, math.Pi/2).Normalize()
		}
	}

	getCur := func(a hexcoord.HexFractional) hexcoord.HexFractional {
		return a.Rotate(origin, math.Pi).Normalize()
	}

	// Test forward direction
	arc := hexcoord.CircularArc{
		I: firstPoint,
		T: getTan(firstPoint),
		E: firstPoint.Rotate(origin, -1.0*sweepRadians).Normalize(),
	}

	curve := arc.Curve()

	rev := float64(1.0)
	if reverse {
		rev = -1.0
	}
	note := rev * sweepRadians

	// Make sure center and sweep are ok
	arcCurve := curve.(hexcoord.ArcCurve)
	assert.True(t, origin.AlmostEquals(arcCurve.Center), fmt.Sprintf("%v (%.3f): Center is not origin.", arc.ToString(), note))
	assertCloseEnough(t, sweepRadians, arcCurve.CentralAngle, fmt.Sprintf("%v (%.3f): Angle size is wrong.", arc.ToString(), note))
	assertCloseEnough(t, rev, arcCurve.Direction, fmt.Sprintf("%v (%.3f): Direction is wrong.", arc.ToString(), note))

	// Test first point.
	assertSample(t, fmt.Sprintf("%v (%.3f)", arc.ToString(), note), 0.0, curve, arc.I, arc.T, getCur(arc.I))
	// Test last point.
	assertSample(t, fmt.Sprintf("%v (%.3f)", arc.ToString(), note), 1.0, curve, arc.E, getTan(arc.E), getCur(arc.E))
	// Test mid point.
	mPos := hexcoord.LerpHexFractional(arc.I, arc.E, 0.5).Normalize()
	assertSample(t, fmt.Sprintf("%v (%.3f)", arc.ToString(), note), 0.5, curve, mPos, getTan(mPos), getCur(mPos))
}

func assertCloseEnough(t *testing.T, a, b float64, msg ...interface{}) {
	if math.Abs(a-b) > 1e-10 {
		assert.Equal(t, a, b, msg...)
	}
}

func TestArcCurve(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	/*testHexes := hexcoord.AreaToSlice(hexcoord.Origin().RingArea(done, 2))
	radii := [...]float64{1.0, 800.0}
	for _, i := range testHexes {
		for _, r := range radii {
			arcCurve(t, r, i.ToHexFractional())
		}
	}*/

	//arcCurve(t, 1.0, hexcoord.Origin().ToHexFractional())
}

func arcCurve(t *testing.T, radius float64, center hexcoord.HexFractional) {

	sampleStep := math.Pi / 4.0
	radV := hexcoord.HexFractionalFromCartesian(1.0, 0.0).Normalize().Multiply(radius)
	//radV := hexcoord.HexFractional{Q: 1.0, R: 1.0}.Normalize().Multiply(radius)
	origin := hexcoord.Origin().ToHexFractional()

	getTan := func(a, center hexcoord.HexFractional) hexcoord.HexFractional {
		return a.Subtract(center).Rotate(origin, math.Pi/2).Normalize()
	}

	for ix := float64(0.0); ix < math.Pi*2; ix = ix + sampleStep {
		for ex := float64(0.0); ex < math.Pi*2; ex = ex + sampleStep {
			if ex == ix {
				continue
			}
			end := radV.Add(center).Rotate(center, ex)
			init := radV.Add(center).Rotate(center, ix)

			initTangent := getTan(init, center)

			scalarCurvature := float64(1.0) / radius

			arc := hexcoord.CircularArc{
				I: init,
				T: initTangent,
				E: end,
			}

			curve := arc.Curve()

			// Test points.
			testCount := 0
			for s := float64(0.0); s <= 1.0; s = s + 0.25 {
				testCount++
				sPoint := radV.Add(center).Rotate(center, lerpFloat(ix, ex, s))

				sTan := getTan(sPoint, center)
				sCurve := center.Subtract(sPoint).Normalize().Multiply(scalarCurvature)
				assertSample(t, arc.ToString(), s, curve, sPoint, sTan, sCurve)
			}

		}
	}
}
