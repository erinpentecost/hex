package curve_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/erinpentecost/hexcoord/curve"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
)

func printArc(c curve.CircularArc) string {
	return fmt.Sprintf("arc(I:%v,%vy; T:%v%v; E:%v%v)", c.I.Q, c.I.R, c.T.Q, c.T.R, c.E.Q, c.E.R)
}

func assertSample(t *testing.T, prefix interface{}, f float64, c curve.Curver, sp, st, sc pos.HexFractional) {

	cp, ct, cc := c.Sample(f)
	if assert.True(t, sp.AlmostEquals(cp), fmt.Sprintf("%v: At sample %v, got position %v but expected %v.", prefix, f, cp.ToString(), sp.ToString())) {
		assert.True(t, st.AlmostEquals(ct), fmt.Sprintf("%v: At sample %v, got tangent %v but expected %v.", prefix, f, ct.ToString(), st.ToString()))
		assert.True(t, sc.AlmostEquals(cc), fmt.Sprintf("%v: At sample %v, got curvature %v but expected %v.", prefix, f, cc.ToString(), sc.ToString()))
	}
}

func TestLineCurve(t *testing.T) {

	done := make(chan interface{})
	defer close(done)

	testHexes := pos.AreaToSlice(pos.Origin().SpiralArea(done, 4))
	origin := pos.Origin().ToHexFractional()

	for _, i := range testHexes {
		for _, e := range testHexes {
			if i == e {
				continue
			}

			tangent := e.ToHexFractional().Subtract(i.ToHexFractional()).Normalize()

			line := curve.CircularArc{
				I: i.ToHexFractional(),
				T: tangent,
				E: e.ToHexFractional(),
			}

			curve := line.Curve()

			assertSample(t, line, 0.0, curve, line.I, line.T, origin)
			assertSample(t, line, 0.1, curve, pos.LerpHexFractional(line.I, line.E, 0.1), line.T, origin)
			assertSample(t, line, 0.5, curve, pos.LerpHexFractional(line.I, line.E, 0.5), line.T, origin)
			assertSample(t, line, 0.75, curve, pos.LerpHexFractional(line.I, line.E, 0.75), line.T, origin)
			assertSample(t, line, 1.0, curve, line.E, line.T, origin)
		}
	}

}

func lerpFloat(a, b, t float64) float64 {
	return a*(1.0-t) + b*t
}

func TestUnitArcCounterClockwise(t *testing.T) {
	originf := pos.Origin().ToHexFractional()
	neighbors := pos.Origin().Neighbors()
	start := pos.HexFractional{Q: 1.0, R: 0.0}
	tan := pos.HexFractional{Q: 1.0, R: -2.0}.Normalize()
	for i, endDiscrete := range neighbors {

		end := endDiscrete.ToHexFractional()
		if end.AlmostEquals(start) {
			continue
		}
		arc := curve.CircularArc{
			E: end,
			I: start,
			T: tan,
		}
		curve := arc.Curve()

		radSwp := float64(i) * math.Pi / 3.0

		endTan := tan.Rotate(originf, radSwp)

		assertCloseEnough(t, radSwp, curve.Length(), "Curve length is wrong.")
		assertSample(t, i, 0.0, curve, start, tan, originf.Subtract(start))
		assertSample(t, i, 1.0, curve, end, endTan, originf.Subtract(end))
	}
}

func TestUnitArcClockwise(t *testing.T) {
	originf := pos.Origin().ToHexFractional()
	neighbors := pos.Origin().Neighbors()
	start := pos.HexFractional{Q: 1.0, R: 0.0}
	tan := pos.HexFractional{Q: -1.0, R: 2.0}.Normalize()
	for i, endDiscrete := range neighbors {

		end := endDiscrete.ToHexFractional()
		if end.AlmostEquals(start) {
			continue
		}
		arc := curve.CircularArc{
			E: end,
			I: start,
			T: tan,
		}
		curve := arc.Curve()

		radSwp := float64(6-i) * math.Pi / 3.0

		endTan := end.Rotate(originf, math.Pi/(-2.0))

		assertCloseEnough(t, radSwp, curve.Length(), "Curve length is wrong.")
		assertSample(t, i, 0.0, curve, start, tan, originf.Subtract(start))
		assertSample(t, i, 1.0, curve, end, endTan, originf.Subtract(end))
	}
}

func assertCloseEnough(t *testing.T, a, b float64, msg ...interface{}) bool {
	if math.Abs(a-b) > 1e-10 {
		return assert.Equal(t, a, b, msg...)
	}
	return true
}
