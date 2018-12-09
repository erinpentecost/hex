package curve_test

import (
	"fmt"
	"testing"

	"github.com/erinpentecost/hexcoord/curve"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
)

func TestBiarcSemiCircle(t *testing.T) {
	up := pos.HexFractional{Q: 1, R: -2}.Normalize()
	down := pos.HexFractional{Q: -1, R: 2}.Normalize()
	biarc := curve.Biarc(
		pos.HexFractional{Q: 0, R: 0},
		up,
		pos.HexFractional{Q: 1, R: 0},
		up,
		1.0)

	a1 := biarc[0]
	a2 := biarc[1]
	c1 := a1.Curve()
	c2 := a2.Curve()

	// Test arc values to make sure they are correct
	assert.True(t, a1.I.AlmostEquals(pos.HexFractional{Q: 0, R: 0}))
	assert.True(t, a1.T.AlmostEquals(up))
	assert.True(t, a2.E.AlmostEquals(pos.HexFractional{Q: 1, R: 0}))
	assert.True(t, a2.T.AlmostEquals(down))

	// Actually sample the arcs at key points now
	end1Point, end1Tangent, _ := c1.Sample(1.0)
	start1Point, start1Tangent, _ := c2.Sample(0.0)
	end2Point, end2Tangent, _ := c2.Sample(1.0)

	// Continuity in the arc
	assert.True(t, end1Tangent.AlmostEquals(start1Tangent))
	assert.True(t, end1Point.AlmostEquals(start1Point))

	// Mid tangent
	assert.True(t, end1Tangent.AlmostEquals(down))

	// Ending values
	assert.True(t, end2Tangent.AlmostEquals(up))
	assert.True(t, end2Point.AlmostEquals(pos.HexFractional{Q: 1, R: 0}))

	// Spin check
	assert.Equal(t, curve.Clockwise, c1.Spin())
	assert.Equal(t, curve.CounterClockwise, c2.Spin())
}

func TestBiarc(t *testing.T) {
	biarc := curve.Biarc(
		pos.HexFractional{Q: 0, R: 0},
		pos.HexFractional{Q: 1, R: -1}.Normalize(),
		pos.HexFractional{Q: 1, R: -1},
		pos.HexFractional{Q: 1, R: 0}.Normalize(),
		1.0)

	a1 := biarc[0]
	a2 := biarc[1]
	c1 := a1.Curve()
	c2 := a2.Curve()

	// Test arc values to make sure they are correct
	assert.True(t, a1.I.AlmostEquals(pos.HexFractional{Q: 0, R: 0}))
	assert.True(t, a1.T.AlmostEquals(pos.HexFractional{Q: 1, R: -1}.Normalize()))
	assert.True(t, a2.E.AlmostEquals(pos.HexFractional{Q: 1, R: -1}))

	// Actually sample the arcs at key points now
	end1Point, end1Tangent, _ := c1.Sample(1.0)
	start1Point, start1Tangent, _ := c2.Sample(0.0)
	end2Point, end2Tangent, _ := c2.Sample(1.0)

	// Continuity in the arc
	assert.True(t, end1Tangent.AlmostEquals(start1Tangent))
	assert.True(t, end1Point.AlmostEquals(start1Point))

	// Ending values
	assert.True(t, end2Tangent.AlmostEquals(pos.HexFractional{Q: 1, R: 0}.Normalize()))
	assert.True(t, end2Point.AlmostEquals(pos.HexFractional{Q: 1, R: -1}))

	// Spin check
	assert.Equal(t, curve.CounterClockwise, c1.Spin())
	assert.Equal(t, curve.Clockwise, c2.Spin())
}

func TestSmoothPathContinuity(t *testing.T) {
	ti := pos.HexFractional{Q: 1, R: -1}.Normalize()
	te := pos.HexFractional{Q: -1, R: 1}.Normalize()
	path := []pos.HexFractional{
		pos.OriginFractional(),
		pos.HexFractional{Q: 1, R: -1},
		pos.HexFractional{Q: 1, R: 0},
		pos.HexFractional{Q: 0, R: 1},
		pos.HexFractional{Q: 1, R: 1},
		pos.HexFractional{Q: 2, R: 0},
		pos.HexFractional{Q: 2, R: -1},
	}

	smoothArcs := curve.SmoothPath(ti, te, path)

	prevArc := smoothArcs[0]
	for _, arc := range smoothArcs[1:] {
		prevCurve := prevArc.Curve()
		curve := arc.Curve()
		p0, p1, _ := prevCurve.Sample(1.0)
		c0, c1, _ := curve.Sample(0.0)
		assert.True(t, p0.AlmostEquals(c0),
			fmt.Sprintf("Sample failed position continuity. Expected %s, got %s.", p0.ToString(), c0.ToString()))
		assert.True(t, p1.AlmostEquals(c1),
			fmt.Sprintf("Sample failed tangent continuity. Expected %s, got %s.", p1.ToString(), c1.ToString()))
	}
}
