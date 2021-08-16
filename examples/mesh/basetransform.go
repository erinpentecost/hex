package mesh

import (
	"github.com/erinpentecost/hex"
	"github.com/erinpentecost/hex/area"
)

var (
	_ Transformer = (*BaseTransform)(nil)
)

type BaseTransform struct {
	area area.Area
}

func (b *BaseTransform) ConvertTo3D(h hex.Hex, actual hex.HexFractional) [3]float32 {
	// ConvertToDetailed3D
	// // glTF defines +Y as up, +Z as forward, and -X as right.
	x, y := actual.ToCartesian()
	z := float32(-0.5)

	// fancy z
	if b.area.ContainsHexes(h) {
		mod := func(a int64) int64 {
			if a < 0 {
				return (a * (-1)) % 2
			}
			return a % 2
		}
		m := mod(h.Q) + 2*mod(h.R)
		switch m {
		case 0:
			z = 0.1
		case 1:
			z = 0.3
		default:
			z = 0
		}
	}

	// put some noise into it
	// TODO: maybe don't put noise in the z direction
	return [3]float32{
		float32(x + (Noise3(x+1000.0, y-3000, float64(z))-0.5)/10),
		z + float32((Noise3(x-9000, y+6000, float64(z))-0.5)/10),
		float32(y + (Noise3(x, y, float64(z))-0.5)/10)}
}

func (b *BaseTransform) HexColor(h hex.Hex) [3]uint8 {
	mod := func(a int64) int64 {
		if a < 0 {
			return (a * (-1)) % 2
		}
		return a % 2
	}

	m := mod(h.Q) + 2*mod(h.R)
	switch m {
	case 0:
		return [3]uint8{255, 222, 222}
	case 1:
		return [3]uint8{222, 255, 222}
	case 2:
		return [3]uint8{222, 222, 255}
	default:
		return [3]uint8{255, 255, 222}
	}
}

func (b *BaseTransform) PointColor(h hex.Hex, direction int) [3]uint8 {
	return [3]uint8{100, 100, 100}
}

func (b *BaseTransform) EdgeColor(h hex.Hex, direction int) (top, bottom [3]uint8) {
	return [3]uint8{190, 190, 190}, [3]uint8{0, 0, 0}
}
