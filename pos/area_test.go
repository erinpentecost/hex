package pos_test

import (
	"testing"

	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
)

func TestAreaSpiralVsHexEqual(t *testing.T) {
	for i := 0; i <= 5; i++ {
		area1 := pos.Origin().SpiralArea(i)
		area2 := pos.Origin().HexArea(i)

		assert.True(t, pos.AreaEqual(area1, area2), "Areas are not equal.")
	}
}

func TestAreaSpiralVsRingEqual(t *testing.T) {
	area1 := pos.Origin().SpiralArea(5)
	area2 := pos.Origin().RingArea(0)
	for i := 0; i <= 5; i++ {
		area2 = pos.AreaUnion(area2, pos.Origin().RingArea(i))
	}

	assert.True(t, pos.AreaEqual(area1, area2), "Areas are not equal.")
}

func TestAreaFlatMap(t *testing.T) {
	line := pos.Origin().LineArea(pos.Hex{
		Q: 14,
		R: 14,
	})

	widenTransform := func(h pos.Hex) pos.Area {
		return h.HexArea(3)
	}

	area1 := pos.AreaFlatMap(line, widenTransform)

	area2 := make([]pos.Hex, 0)

	go func() {
		for _, h := range line {
			wide := widenTransform(h)
			area2 = append(area2, wide...)
		}
	}()

	assert.True(t, pos.AreaEqual(area1, area2), "Areas are not equal.")

}

func TestAreaEqual(t *testing.T) {
	area1 := pos.Origin().RingArea(1)
	area2 := pos.Origin().RingArea(1)
	area3 := pos.Origin().RingArea(1)
	area4 := pos.Origin().RingArea(2)

	assert.True(t, pos.AreaEqual(area1, area2), "Areas are not equal.")
	assert.False(t, pos.AreaEqual(area4, area3), "Areas are equal.")
}

func TestAreaIntersection(t *testing.T) {

	assert.True(t,
		pos.AreaEqual(pos.Origin().HexArea(10), pos.Origin().HexArea(10)),
		"Areas are not equal.")

	identity := pos.AreaIntersection(
		pos.Origin().HexArea(10),
		pos.Origin().HexArea(10))

	assert.True(t,
		pos.AreaEqual(pos.Origin().HexArea(10), identity),
		"Intersection failed on matched input.")

	ringCheck := pos.AreaIntersection(
		pos.Origin().RingArea(4),
		pos.Origin().HexArea(10))

	assert.True(t,
		pos.AreaEqual(ringCheck, pos.Origin().RingArea(4)),
		"Intersection failed with unmatched input.")
}
