package mesh

import "github.com/erinpentecost/hexcoord/pos"

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

func (b *BaseTransform) EmbedNormals() bool {
	return false
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

	return [3]float32{float32(x), float32(y), z}
}
