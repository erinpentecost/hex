package hex

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func maxInt(a, k int64) int64 {
	if a > k {
		return a
	}
	return k
}

func minInt(a, k int64) int64 {
	if a < k {
		return a
	}
	return k
}

// HexArea returns the set of hexes that form a larger hex area
// centered around the starting hex and with the given radius.
// The order of elements returned is not set.
// A radius of 0 will return the center hex.
func HexArea(h Hex, radius int64) []Hex {
	area := make([]Hex, 0)
	for q := -1 * radius; q <= radius; q++ {
		r1 := maxInt(-1*radius, -1*(q+radius))
		r2 := minInt(radius, (-1*q)+radius)

		for r := r1; r <= r2; r++ {
			area = append(area, Hex{
				Q: q,
				R: r,
			})
		}
	}
	return area
}

func TestHexHashIdentity(t *testing.T) {
	o1 := Origin()
	o2 := Origin()
	assert.Equal(t, o1, o2, "Origin copy is not equal.")

	p1 := Hex{
		Q: 10,
		R: -8888888,
	}
	p2 := Hex{
		Q: 10,
		R: -8888888,
	}
	assert.Equal(t, p1, p2, "Hex copy is not equal.")
}

func TestHexAdd(t *testing.T) {
	h1 := Hex{
		Q: 5,
		R: 5,
	}
	h2 := Hex{
		Q: 9,
		R: 9,
	}
	hsum := h1.Add(h2)
	hexpected := Hex{
		Q: 14,
		R: 14,
	}

	assert.Equal(t, hexpected, hsum, "Hex add is not correct.")
}

func TestHexDistance(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	closeHexes := HexArea(Origin(), 1)
	for _, h := range closeHexes {
		if h == Origin() {
			assert.EqualValues(t, 0, h.DistanceTo(Origin()))
		} else {
			assert.EqualValues(t, 1, h.DistanceTo(Origin()), fmt.Sprintf("Hex distance to %v is wrong.", h))
		}
	}
}

func testDirectionEquality(t *testing.T, testOrigin Hex) {
	for a := int64(-9); a < 9; a++ {
		if a == 0 {
			continue
		}
		for d := -9; d < 9; d++ {
			dh := Direction(d).Multiply(a).Add(testOrigin)

			rh := Direction(3 + d).Multiply(-1 * a).Add(testOrigin)

			assert.Equal(t, dh, rh, "Reversed distance hexes are not equal.")

			oh := Direction(3 + d).Multiply(a).Add(testOrigin)

			assert.NotEqual(t, dh, oh, fmt.Sprintf("Opposite hexes are equal with d=%v.", d))

			assert.Equal(t, 2*testOrigin.DistanceTo(oh), dh.DistanceTo(oh), "Distance is not expected for opposite hexes.")
		}
	}
}

func TestDirectionEquality(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	testHexes := HexArea(Origin(), 10)
	for _, h := range testHexes {
		testDirectionEquality(t, h)
	}
}

func TestFractionalConversion(t *testing.T) {
	testHexes := HexArea(Origin(), 10)
	for _, h := range testHexes {
		frac := h.ToHexFractional()
		recast := frac.ToHex()
		assert.Equal(t, h, recast)
	}
}

func TestLine(t *testing.T) {
	segs := [][]Hex{
		{{Q: 1, R: -2},
			{Q: 1, R: -1},
			{Q: 1, R: 0},
			{Q: 1, R: 1}},
		{{Q: 1, R: 1},
			{Q: 0, R: 1},
			{Q: -1, R: 1},
			{Q: -2, R: 1}},
		{{Q: -2, R: 1},
			{Q: -1, R: 0},
			{Q: 0, R: -1},
			{Q: 1, R: -2}},
		{{Q: -1, R: 0},
			{Q: 0, R: 0},
			{Q: 1, R: 0}},
	}

	for _, seg := range segs {
		first := seg[0]
		last := seg[len(seg)-1]
		t.Run(fmt.Sprintf("from-%s-to-%s", first.String(), last.String()), func(t *testing.T) {
			found := first.LineTo(last)
			assert.Equal(t, seg, found)
		})
		t.Run(fmt.Sprintf("from-%s-to-%s", last.String(), first.String()), func(t *testing.T) {
			found := last.LineTo(first)
			for i, s := range seg {
				assert.Equal(t, s, found[len(seg)-i-1])
			}
		})
	}
}
