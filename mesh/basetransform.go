package mesh

import (
	"github.com/erinpentecost/hexcoord/pos"
)

var (
	_ OptimizedTransformer = (*BaseTransform)(nil)
	_ DetailedTransformer  = (*BaseTransform)(nil)
)

type BaseTransform struct{}

func (b *BaseTransform) ConvertToOptimized3D(actual pos.HexFractional) [3]float32 {
	x, y := actual.ToCartesian()
	// ConvertToDetailed3D
	// // glTF defines +Y as up, +Z as forward, and -X as right.
	return [3]float32{float32(x), 0.0, float32(y)}
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

func (b *BaseTransform) EdgeColor(h, n1, n2 pos.Hex) [3]uint8 {
	return [3]uint8{100, 100, 100}
}

func (b *BaseTransform) ConvertToDetailed3D(hd pos.Hex, actual pos.HexFractional) [3]float32 {
	x, y := actual.ToCartesian()

	// fancy z
	z := float32(0)
	mod := func(a int64) int64 {
		if a < 0 {
			return (a * (-1)) % 2
		}
		return a % 2
	}
	m := mod(hd.Q) + 2*mod(hd.R)
	switch m {
	case 0:
		z = 1
	case 1:
		z = 2
	case 2:
		z = 3
	default:
		z = 0
	}

	return [3]float32{float32(x), z, float32(y)}
}
