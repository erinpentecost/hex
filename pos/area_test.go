package pos_test

import (
	"testing"

	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
)

func TestAreaSpiralVsHexEqual(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	for i := 0; i <= 5; i++ {
		area1 := pos.Origin().SpiralArea(done, i)
		area2 := pos.Origin().HexArea(done, i)

		assert.True(t, pos.AreaEqual(area1, area2), "Areas are not equal.")
	}
}

func TestAreaSpiralVsRingEqual(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	area1 := pos.Origin().SpiralArea(done, 5)
	area2 := pos.Origin().RingArea(done, 0)
	for i := 0; i <= 5; i++ {
		area2 = pos.AreaUnion(done, area2, pos.Origin().RingArea(done, i))
	}

	assert.True(t, pos.AreaEqual(area1, area2), "Areas are not equal.")
}

func TestAreaSum(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	area1 := pos.Origin().SpiralArea(done, 5)

	area2 := make(chan (<-chan pos.Hex))

	go func() {
		defer close(area2)

		for i := 0; i <= 5; i++ {
			area2 <- pos.Origin().RingArea(done, i)
		}
	}()

	assert.True(t, pos.AreaEqual(area1, pos.AreaSum(done, area2)), "Areas are not equal.")
}

func TestAreaFlatMap(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	line := pos.AreaToSlice(pos.Origin().LineArea(done, pos.Hex{
		Q: 14,
		R: 14,
	}))

	widenTransform := func(d <-chan interface{}, h pos.Hex) <-chan pos.Hex {
		return h.HexArea(d, 3)
	}

	area1 := pos.AreaFlatMap(done, pos.Area(line...), widenTransform)

	area2 := make(chan pos.Hex)

	go func() {
		defer close(area2)
		for _, h := range line {
			wide := widenTransform(done, h)
			for wh := range wide {
				area2 <- wh
			}
		}
	}()

	assert.True(t, pos.AreaEqual(area1, area2), "Areas are not equal.")

}

func TestAreaEqual(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	area1 := pos.Origin().RingArea(done, 1)
	area2 := pos.Origin().RingArea(done, 1)
	area3 := pos.Origin().RingArea(done, 1)
	area4 := pos.Origin().RingArea(done, 2)

	assert.True(t, pos.AreaEqual(area1, area2), "Areas are not equal.")
	assert.False(t, pos.AreaEqual(area4, area3), "Areas are equal.")
}

func TestAreaIntersection(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	assert.True(t,
		pos.AreaEqual(pos.Origin().HexArea(done, 10), pos.Origin().HexArea(done, 10)),
		"Areas are not equal.")

	identity := pos.AreaIntersection(done,
		pos.Origin().HexArea(done, 10),
		pos.Origin().HexArea(done, 10))

	assert.True(t,
		pos.AreaEqual(pos.Origin().HexArea(done, 10), identity),
		"Intersection failed on matched input.")

	ringCheck := pos.AreaIntersection(done,
		pos.Origin().RingArea(done, 4),
		pos.Origin().HexArea(done, 10))

	assert.True(t,
		pos.AreaEqual(ringCheck, pos.Origin().RingArea(done, 4)),
		"Intersection failed with unmatched input.")
}
