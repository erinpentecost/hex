package pos_test

import (
	"fmt"
	"testing"

	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
)

func TestHexHashIdentity(t *testing.T) {
	o1 := pos.Origin()
	o2 := pos.Origin()
	assert.Equal(t, o1, o2, "Origin copy is not equal.")

	p1 := pos.Hex{
		Q: 10,
		R: -8888888,
	}
	p2 := pos.Hex{
		Q: 10,
		R: -8888888,
	}
	assert.Equal(t, p1, p2, "Hex copy is not equal.")
}

func TestHexAdd(t *testing.T) {
	h1 := pos.Hex{
		Q: 5,
		R: 5,
	}
	h2 := pos.Hex{
		Q: 9,
		R: 9,
	}
	hsum := h1.Add(h2)
	hexpected := pos.Hex{
		Q: 14,
		R: 14,
	}

	assert.Equal(t, hexpected, hsum, "Hex add is not correct.")
}

func TestHexDistance(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	closeHexes := pos.Origin().HexArea(done, 1)
	for h := range closeHexes {
		if h == pos.Origin() {
			assert.Equal(t, 0, h.DistanceTo(pos.Origin()))
		} else {
			assert.Equal(t, 1, h.DistanceTo(pos.Origin()), fmt.Sprintf("Hex distance to %v is wrong.", h))
		}
	}
}

func testDirectionEquality(t *testing.T, testOrigin pos.Hex) {
	for a := -9; a < 9; a++ {
		if a == 0 {
			continue
		}
		for d := -9; d < 9; d++ {
			dh := pos.Direction(d).Multiply(a).Add(testOrigin)

			rh := pos.Direction(3 + d).Multiply(-1 * a).Add(testOrigin)

			assert.Equal(t, dh, rh, "Reversed distance hexes are not equal.")

			oh := pos.Direction(3 + d).Multiply(a).Add(testOrigin)

			assert.NotEqual(t, dh, oh, fmt.Sprintf("Opposite hexes are equal with d=%v.", d))

			assert.Equal(t, 2*testOrigin.DistanceTo(oh), dh.DistanceTo(oh), "Distance is not expected for opposite hexes.")
		}
	}
}

func TestDirectionEquality(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	testHexes := pos.Origin().HexArea(done, 10)
	for h := range testHexes {
		testDirectionEquality(t, h)
	}
}

func TestFractionalConversion(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	testHexes := pos.Origin().HexArea(done, 10)
	for h := range testHexes {
		frac := h.ToHexFractional()
		recast := frac.ToHex()
		assert.Equal(t, h, recast)
	}
}
