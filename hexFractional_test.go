package hexcoord_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/erinpentecost/hexcoord"
	"github.com/stretchr/testify/assert"
)

func TestHexFractionalHashIdentity(t *testing.T) {
	p1 := hexcoord.HexFractional{
		Q: 10.0,
		R: -888.8888,
	}
	p2 := hexcoord.HexFractional{
		Q: 10.0,
		R: -888.8888,
	}
	assert.Equal(t, p1, p2, "Hex copy is not equal.")
}

func TestHexFractionalLength(t *testing.T) {

	tests := []struct {
		H hexcoord.HexFractional
		E float64
	}{
		{H: hexcoord.HexFractional{Q: 0, R: -1}, E: 1},
		{H: hexcoord.HexFractional{Q: 1, R: -1}, E: 1},
		{H: hexcoord.HexFractional{Q: 1, R: 0}, E: 1},
		{H: hexcoord.HexFractional{Q: 0, R: 1}, E: 1},
		{H: hexcoord.HexFractional{Q: -1, R: 1}, E: 1},
		{H: hexcoord.HexFractional{Q: -1, R: 0}, E: 1},

		{H: hexcoord.HexFractional{Q: 1, R: -2}, E: math.Sqrt(3)},
		{H: hexcoord.HexFractional{Q: 2, R: -1}, E: math.Sqrt(3)},
		{H: hexcoord.HexFractional{Q: 1, R: 1}, E: math.Sqrt(3)},
		{H: hexcoord.HexFractional{Q: -1, R: 2}, E: math.Sqrt(3)},
		{H: hexcoord.HexFractional{Q: -2, R: 1}, E: math.Sqrt(3)},
		{H: hexcoord.HexFractional{Q: -1, R: -1}, E: math.Sqrt(3)},

		{H: hexcoord.HexFractional{Q: 0, R: -2}, E: 2},
		{H: hexcoord.HexFractional{Q: -2, R: 2}, E: 2},
	}

	for _, he := range tests {
		assert.Equal(t, he.E, he.H.Length(), fmt.Sprintf("HexFractional distance to %v is wrong.", he.H))
	}
}

func TestHexFractionalNormalize(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	testHexes := hexcoord.HexOrigin().HexArea(done, 10)
	for h := range testHexes {
		len := h.ToHexFractional().Normalize().Length()
		assert.InEpsilonf(t, 1.0, 0.0000001, len, fmt.Sprintf("HexFractional normalization for %v is wrong.", h))
	}

}

func TestCartesian(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	testHexes := hexcoord.HexOrigin().HexArea(done, 10)
	for h := range testHexes {
		hf := h.ToHexFractional()
		converted := hexcoord.HexFractionalFromCartesian(hf.ToCartesian())
		assert.True(t, hf.AlmostEquals(converted), fmt.Sprintf("Expected %v, got %v.", hf, converted))
	}
}

func TestRotate(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	radianStep := float64(math.Pi / 3.0)

	testHexes := hexcoord.HexOrigin().SpiralArea(done, 10)
	for h := range testHexes {
		hf := h.ToHexFractional()
		for i, n := range h.Neighbors() {
			nfe := n.ToHexFractional()
			nft := h.Neighbor(0).ToHexFractional().Rotate(hf, float64(i)*radianStep)
			assert.True(t, nfe.AlmostEquals(nft),
				fmt.Sprintf("Rotated %v about %v by %v˚. Expected %v, got %v.", h.Neighbor(0), hf, i*60, nfe, nft))
		}
	}
}

func TestAngleTo(t *testing.T) {
	sqrt3 := math.Sqrt(3.0)
	o := hexcoord.HexOrigin()

	for ia, ra := range o.Neighbors() {
		for ib, rb := range o.Neighbors() {
			diff := ia - ib
			assert.Equal(t,
				sqrt3*(float64(diff)),
				ra.ToHexFractional().AngleTo(rb.ToHexFractional()),
				fmt.Sprintf("Angle from %v to %v (offset by %v) is wrong.", ra, rb, diff))
		}
	}
}
