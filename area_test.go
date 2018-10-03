package hexcoord_test

import (
	"testing"

	"github.com/erinpentecost/hexcoord"
	"github.com/stretchr/testify/assert"
)

func TestAreaSpiralVsHexEqual(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	for i := 0; i <= 5; i++ {
		area1 := hexcoord.Origin().SpiralArea(done, i)
		area2 := hexcoord.Origin().HexArea(done, i)

		assert.True(t, hexcoord.AreaEqual(area1, area2), "Areas are not equal.")
	}
}

func TestAreaSpiralVsRingEqual(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	area1 := hexcoord.Origin().SpiralArea(done, 5)
	area2 := hexcoord.Origin().RingArea(done, 0)
	for i := 0; i <= 5; i++ {
		area2 = hexcoord.AreaUnion(done, area2, hexcoord.Origin().RingArea(done, i))
	}

	assert.True(t, hexcoord.AreaEqual(area1, area2), "Areas are not equal.")
}

func TestAreaSum(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	area1 := hexcoord.Origin().SpiralArea(done, 5)

	area2 := make(chan (<-chan hexcoord.Hex))

	go func() {
		defer close(area2)

		for i := 0; i <= 5; i++ {
			area2 <- hexcoord.Origin().RingArea(done, i)
		}
	}()

	assert.True(t, hexcoord.AreaEqual(area1, hexcoord.AreaSum(done, area2)), "Areas are not equal.")
}

func TestAreaFlatMap(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	line := hexcoord.AreaToSlice(hexcoord.Origin().LineArea(done, hexcoord.Hex{
		Q: 14,
		R: 14,
	}))

	widenTransform := func(d <-chan interface{}, h hexcoord.Hex) <-chan hexcoord.Hex {
		return h.HexArea(d, 3)
	}

	area1 := hexcoord.AreaFlatMap(done, hexcoord.Area(line...), widenTransform)

	area2 := make(chan hexcoord.Hex)

	go func() {
		defer close(area2)
		for _, h := range line {
			wide := widenTransform(done, h)
			for wh := range wide {
				area2 <- wh
			}
		}
	}()

	assert.True(t, hexcoord.AreaEqual(area1, area2), "Areas are not equal.")

}

func TestAreaEqual(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	area1 := hexcoord.Origin().RingArea(done, 1)
	area2 := hexcoord.Origin().RingArea(done, 1)
	area3 := hexcoord.Origin().RingArea(done, 1)
	area4 := hexcoord.Origin().RingArea(done, 2)

	assert.True(t, hexcoord.AreaEqual(area1, area2), "Areas are not equal.")
	assert.False(t, hexcoord.AreaEqual(area4, area3), "Areas are equal.")
}

func TestAreaIntersection(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	assert.True(t,
		hexcoord.AreaEqual(hexcoord.Origin().HexArea(done, 10), hexcoord.Origin().HexArea(done, 10)),
		"Areas are not equal.")

	identity := hexcoord.AreaIntersection(done,
		hexcoord.Origin().HexArea(done, 10),
		hexcoord.Origin().HexArea(done, 10))

	assert.True(t,
		hexcoord.AreaEqual(hexcoord.Origin().HexArea(done, 10), identity),
		"Intersection failed on matched input.")

	ringCheck := hexcoord.AreaIntersection(done,
		hexcoord.Origin().RingArea(done, 4),
		hexcoord.Origin().HexArea(done, 10))

	assert.True(t,
		hexcoord.AreaEqual(ringCheck, hexcoord.Origin().RingArea(done, 4)),
		"Intersection failed with unmatched input.")
}
