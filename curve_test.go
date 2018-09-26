package hexcoord_test

import (
	"fmt"
	"testing"

	"github.com/erinpentecost/hexcoord"
	"github.com/stretchr/testify/assert"
)

func printArc(c hexcoord.CircularArc) string {
	return fmt.Sprintf("arc(I:%v,%vy; T:%v%v; E:%v%v)", c.I.Q, c.I.R, c.T.Q, c.T.R, c.E.Q, c.E.R)
}

func assertSample(t *testing.T, f float64, c hexcoord.CurveSegmenter, sp, st, sc hexcoord.HexFractional) {

	cp, ct, cc := c.Sample(f)
	assert.True(t, sp.AlmostEquals(cp), fmt.Sprintf("At sample %v, got position %v but expected %v.", f, cp, sp))
	assert.True(t, st.AlmostEquals(ct), fmt.Sprintf("At sample %v, got tangent %v but expected %v.", f, ct, st))
	assert.True(t, sc.AlmostEquals(cc), fmt.Sprintf("At sample %v, got curvature %v but expected %v.", f, cc, sc))
}

func TestLineCurve(t *testing.T) {

	done := make(chan interface{})
	defer close(done)

	testHexes := hexcoord.AreaToSlice(hexcoord.HexOrigin().SpiralArea(done, 2))
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
