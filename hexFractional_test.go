package hex

import (
	"fmt"
	"math"
	"testing"

	"github.com/erinpentecost/hex/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHexFractionalHashIdentity(t *testing.T) {
	p1 := HexFractional{
		Q: 10.0,
		R: -888.8888,
	}
	p2 := HexFractional{
		Q: 10.0,
		R: -888.8888,
	}
	assert.Equal(t, p1, p2, "Hex copy is not equal.")
}

func TestHexFractionalLength(t *testing.T) {

	tests := []struct {
		H HexFractional
		E float64
	}{
		{H: HexFractional{Q: 0, R: -1}, E: 1},
		{H: HexFractional{Q: 1, R: -1}, E: 1},
		{H: HexFractional{Q: 1, R: 0}, E: 1},
		{H: HexFractional{Q: 0, R: 1}, E: 1},
		{H: HexFractional{Q: -1, R: 1}, E: 1},
		{H: HexFractional{Q: -1, R: 0}, E: 1},

		{H: HexFractional{Q: 1, R: -2}, E: math.Sqrt(3)},
		{H: HexFractional{Q: 2, R: -1}, E: math.Sqrt(3)},
		{H: HexFractional{Q: 1, R: 1}, E: math.Sqrt(3)},
		{H: HexFractional{Q: -1, R: 2}, E: math.Sqrt(3)},
		{H: HexFractional{Q: -2, R: 1}, E: math.Sqrt(3)},
		{H: HexFractional{Q: -1, R: -1}, E: math.Sqrt(3)},

		{H: HexFractional{Q: 0, R: -2}, E: 2},
		{H: HexFractional{Q: -2, R: 2}, E: 2},
	}

	for _, he := range tests {
		assert.Equal(t, he.E, he.H.Length(), fmt.Sprintf("HexFractional distance to %v is wrong.", he.H))
	}
}

func TestHexFractionalNormalize(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	testHexes := HexArea(Origin(), 10)
	for _, h := range testHexes {
		if h == Origin() {
			continue
		}
		len := h.ToHexFractional().Normalize().Length()
		assert.InEpsilonf(t, 1.0, len, 0.0000001, fmt.Sprintf("HexFractional normalization for %v is wrong.", h))
	}

}

func TestCartesian(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	testHexes := HexArea(Origin(), 10)
	for _, h := range testHexes {
		hf := h.ToHexFractional()
		converted := HexFractionalFromCartesian(hf.ToCartesian())
		assert.True(t, hf.AlmostEquals(converted), fmt.Sprintf("Expected %v, got %v.", hf, converted))
	}

	ox, oy := Origin().ToHexFractional().ToCartesian()
	assert.Equal(t, 0.0, ox, "Origin x is wrong.")
	assert.Equal(t, 0.0, oy, "Origin y is wrong.")
}

// avoid compiler optimizations
var benchx, benchy float64

func BenchmarkToCartesian(b *testing.B) {
	hexf := HexFractional{Q: 237.0, R: -23455.0}
	for i := 0; i < b.N; i++ {
		benchx, benchy = hexf.ToCartesian()
	}
}

// avoid compiler optimizations
var benchHex HexFractional

func BenchmarkFromCartesian(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchHex = HexFractionalFromCartesian(2345.0, 562.5)
	}
}

func TestRotate(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	radianStep := float64(math.Pi / 3.0)

	testHexes := HexArea(Origin(), 10)
	for _, h := range testHexes {
		hf := h.ToHexFractional()
		for i, n := range h.Neighbors() {
			nfe := n.ToHexFractional()
			nft := h.Neighbor(0).ToHexFractional().Rotate(hf, float64(i)*radianStep)
			assert.True(t, nfe.AlmostEquals(nft),
				fmt.Sprintf("Rotated %v about %v by %v˚. Expected %v, got %v.", h.Neighbor(0), hf, i*60, nfe, nft))
		}
	}
}

func TestOpposite(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	testHexes := Origin().Neighbors()
	for _, h := range testHexes {
		hf := h.ToHexFractional()
		assert.True(t, hf.Rotate(OriginFractional(), math.Pi).AlmostEquals(hf.Multiply(-1.0)))
	}
}

func TestScale(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	testHexes := HexArea(Hex{Q: 9999, R: 664}, 4)
	scalar := []float64{-1000.0, 1.0, 3.0}

	for _, a := range testHexes {
		af := a.ToHexFractional()
		ax, ay := af.ToCartesian()
		for _, b := range testHexes {
			bf := b.ToHexFractional()
			bx, by := bf.ToCartesian()
			for _, s := range scalar {
				// hex scaling
				hScale := af.Multiply(s).Add(bf).Multiply(s)
				// cartesian scaling
				scale := func(a, b, s float64) float64 {
					return (a*s + b) * s
				}
				cScaleX := scale(ax, bx, s)
				cScaleY := scale(ay, by, s)
				cScale := HexFractionalFromCartesian(cScaleX, cScaleY)

				require.True(t, hScale.AlmostEquals(cScale), fmt.Sprintf("hex derived %v is not cartesian derived %v", hScale.String(), cScale.String()))
			}
		}
	}
}

func TestAngleTo(t *testing.T) {

	o := Origin()

	pid3 := math.Pi / 3.0
	toRad := func(a, b int) float64 {
		// get inner angle at all times
		rot := ((a % 6) - (b % 6)) % 6
		if rot < 0 {
			rot = rot * (-1)
		}
		if rot == 4 {
			rot = 2
		}
		if rot == 5 {
			rot = 1
		}

		// convert to rads
		return pid3 * (float64(rot))
	}

	for ia, ra := range o.Neighbors() {
		for ib, rb := range o.Neighbors() {
			assert.True(t,
				internal.CloseEnough(toRad(ia, ib), ra.ToHexFractional().AngleTo(rb.ToHexFractional())),
				fmt.Sprintf("Angle from %v to %v (offset by %v) is wrong.", ra, rb, ia-ib))
		}
	}
}
