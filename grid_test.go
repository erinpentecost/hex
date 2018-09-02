package hexcoord_test

import (
	"fmt"
	"testing"

	"github.com/erinpentecost/hexcoord"
	"github.com/stretchr/testify/assert"
)

func TestHexHashIdentity(t *testing.T) {
	o1 := hexcoord.HexOrigin()
	o2 := hexcoord.HexOrigin()
	assert.Equal(t, o1, o2, "Origin copy is not equal.")
}

func TestHexAdd(t *testing.T) {
	h1 := hexcoord.Hex{
		Q: 5,
		R: 5,
	}
	h2 := hexcoord.Hex{
		Q: 9,
		R: 9,
	}
	hsum := h1.Add(h2)
	hexpected := hexcoord.Hex{
		Q: 14,
		R: 14,
	}

	assert.Equal(t, hexpected, hsum, "Hex add is not correct.")
}

func TestHexDistance(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	closeHexes := hexcoord.HexOrigin().HexArea(done, 1)
	for h := range closeHexes {
		if h == hexcoord.HexOrigin() {
			assert.Equal(t, 0, h.DistanceTo(hexcoord.HexOrigin()))
		} else {
			assert.Equal(t, 1, h.DistanceTo(hexcoord.HexOrigin()), fmt.Sprintf("Hex distance to %v is wrong.", h))
		}
	}
}

func testDirectionEquality(t *testing.T, testOrigin hexcoord.Hex) {
	for a := -9; a < 9; a++ {
		if a == 0 {
			continue
		}
		for d := -9; d < 9; d++ {
			dh := hexcoord.HexDirection(d).Multiply(a).Add(testOrigin)

			rh := hexcoord.HexDirection(3 + d).Multiply(-1 * a).Add(testOrigin)

			assert.Equal(t, dh, rh, "Reversed distance hexes are not equal.")

			oh := hexcoord.HexDirection(3 + d).Multiply(a).Add(testOrigin)

			assert.NotEqual(t, dh, oh, fmt.Sprintf("Opposite hexes are equal with d=%v.", d))

			assert.Equal(t, 2*testOrigin.DistanceTo(oh), dh.DistanceTo(oh), "Distance is not expected for opposite hexes.")
		}
	}
}

func TestDirectionEquality(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	testHexes := hexcoord.HexOrigin().HexArea(done, 10)
	for h := range testHexes {
		testDirectionEquality(t, h)
	}
}

func TestFractionalConversion(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	testHexes := hexcoord.HexOrigin().HexArea(done, 10)
	for h := range testHexes {
		frac := h.ToHexFractional()
		recast := frac.ToHex()
		assert.Equal(t, h, recast)
	}
}

func TestAreaSpiralVsHexEqual(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	for i := 0; i <= 5; i++ {
		area1 := hexcoord.HexOrigin().SpiralArea(done, i)
		area2 := hexcoord.HexOrigin().HexArea(done, i)

		assert.True(t, hexcoord.AreaEqual(area1, area2), "Areas are not equal.")
	}
}

func TestAreaSpiralVsRingEqual(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	area1 := hexcoord.HexOrigin().SpiralArea(done, 5)
	area2 := hexcoord.HexOrigin().RingArea(done, 0)
	for i := 0; i <= 5; i++ {
		area2 = hexcoord.AreaUnion(done, area2, hexcoord.HexOrigin().RingArea(done, i))
	}

	assert.True(t, hexcoord.AreaEqual(area1, area2), "Areas are not equal.")
}

func TestAreaEqual(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	area1 := hexcoord.HexOrigin().RingArea(done, 1)
	area2 := hexcoord.HexOrigin().RingArea(done, 1)
	area3 := hexcoord.HexOrigin().RingArea(done, 1)
	area4 := hexcoord.HexOrigin().RingArea(done, 2)

	assert.True(t, hexcoord.AreaEqual(area1, area2), "Areas are not equal.")
	assert.False(t, hexcoord.AreaEqual(area4, area3), "Areas are equal.")
}

func TestAreaIntersection(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	assert.True(t,
		hexcoord.AreaEqual(hexcoord.HexOrigin().HexArea(done, 10), hexcoord.HexOrigin().HexArea(done, 10)),
		"Areas are not equal.")

	identity := hexcoord.AreaIntersection(done,
		hexcoord.HexOrigin().HexArea(done, 10),
		hexcoord.HexOrigin().HexArea(done, 10))

	assert.True(t,
		hexcoord.AreaEqual(hexcoord.HexOrigin().HexArea(done, 10), identity),
		"Intersection failed on matched input.")

	ringCheck := hexcoord.AreaIntersection(done,
		hexcoord.HexOrigin().RingArea(done, 4),
		hexcoord.HexOrigin().HexArea(done, 10))

	assert.True(t,
		hexcoord.AreaEqual(ringCheck, hexcoord.HexOrigin().RingArea(done, 4)),
		"Intersection failed with unmatched input.")
}
