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

func assertSample(t *testing.T, f float64, c hexcoord.Curver, sp, st, sc hexcoord.HexFractional) {

	cp, ct, cc := c.Sample(f)
	assert.True(t, sp.AlmostEquals(cp), fmt.Sprintf("At sample %v, got position %v but expected %v.", f, cp, sp))
	assert.True(t, st.AlmostEquals(ct), fmt.Sprintf("At sample %v, got tangent %v but expected %v.", f, ct, st))
	assert.True(t, sc.AlmostEquals(cc), fmt.Sprintf("At sample %v, got curvature %v but expected %v.", f, cc, sc))
}

func TestLineCurve(t *testing.T) {

	done := make(chan interface{})
	defer close(done)

	testHexes := hexcoord.AreaToSlice(hexcoord.HexOrigin().SpiralArea(done, 4))
	origin := hexcoord.HexOrigin().ToHexFractional()

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

			assertSample(t, 0.0, curve, line.I, line.T, origin)
			assertSample(t, 0.1, curve, hexcoord.LerpHexFractional(line.I, line.E, 0.1), line.T, origin)
			assertSample(t, 0.5, curve, hexcoord.LerpHexFractional(line.I, line.E, 0.5), line.T, origin)
			assertSample(t, 0.75, curve, hexcoord.LerpHexFractional(line.I, line.E, 0.75), line.T, origin)
			assertSample(t, 1.0, curve, line.E, line.T, origin)
		}
	}

}

func lerpFloat(a, b, t float64) float64 {
	return a*(1.0-t) + b*t
}

func TestArcCurve(t *testing.T) {

	done := make(chan interface{})
	defer close(done)

	testHexes := hexcoord.AreaToSlice(hexcoord.HexOrigin().SpiralArea(done, 4))
	for _, i := range testHexes {
		for r := float64(1.0); r < 3.0; r = r + 0.5 {
			arcCurve(t, r, i.ToHexFractional())
		}
	}
}

func arcCurve(t *testing.T, radius float64, center hexcoord.HexFractional) {

	sampleStep := math.Pi / 5.0
	radV := hexcoord.HexFractional{Q: 1.0, R: 1.0}.Multiply(radius)

	for ex := float64(0.0); ex < math.Pi*2; ex = ex + sampleStep {
		for ix := float64(0.0); ix < math.Pi*2; ix = ix + sampleStep {
			clockwise := ix < ex
			end := radV.Add(center).Rotate(center, ex)
			init := radV.Add(center).Rotate(center, ix)

			initTangent := center.Subtract(init).Normalize()

			scalarCurvature := float64(1.0) / radius

			arc := hexcoord.CircularArc{
				I: init,
				T: initTangent,
				E: end,
			}

			curve := arc.Curve()

			// Test points.
			for s := float64(0.0); s <= 1.0; s = s + 0.25 {
				sPoint := radV.Add(center).Rotate(center, lerpFloat(ix, ex, s))
				tangentLine := hexcoord.HexFractional{
					Q: -1 * sPoint.R,
					R: sPoint.Q,
				}
				dir := float64(1.0)
				if clockwise {
					dir = -1.0
				}
				sTan := sPoint.Rotate(center, dir).ProjectOn(tangentLine).Normalize()
				sCurve := center.Subtract(sPoint).Normalize().Multiply(scalarCurvature)
				assertSample(t, 0.1, curve, sPoint, sTan, sCurve)
			}

		}
	}
}
