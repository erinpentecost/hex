package mesh

import (
	"github.com/erinpentecost/hexcoord/pos"
)

var (
	_ Transformer = (*BaseTransform)(nil)
)

type BaseTransform struct{}

func (b *BaseTransform) ConvertTo3D(h *pos.Hex, actual pos.HexFractional) [3]float32 {
	// ConvertToDetailed3D
	// // glTF defines +Y as up, +Z as forward, and -X as right.
	x, y := actual.ToCartesian()

	if h == nil {
		return [3]float32{float32(x), 0.0, float32(y)}
	}

	// fancy z
	z := float32(0)
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

	return [3]float32{float32(x), z, float32(y)}
}

func (b *BaseTransform) HexColor(h pos.Hex) [3]uint8 {
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

func (b *BaseTransform) PointColor(h pos.Hex, direction int) [3]uint8 {
	return [3]uint8{100, 100, 100}
}

func (b *BaseTransform) EdgeColor(h pos.Hex, direction int) [3]uint8 {
	return [3]uint8{150, 150, 150}
}
