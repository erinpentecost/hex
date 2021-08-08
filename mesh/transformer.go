package mesh

import (
	"github.com/erinpentecost/hexcoord/pos"
)

// Transformer converts 2-dimensional cartesian space into three dimensions.
// Use this to select which dimension is "up" and do stretching if needed.
type Transformer interface {
	// ConvertTo3D converts some hex vector to 3D cartesian space.
	//
	// h is nil when used with EncodeOptimized, since the owner of the point
	// is indeterminate.
	//
	// h is nil when called with intermediate values on the faces of the edge
	// rectangles between hexes when used with EncodeDetailed.
	//   TODO: this might make this realllly hard
	//
	// glTF defines +Y as up, +Z as forward, and -X as right.
	ConvertTo3D(h *pos.Hex, actual pos.HexFractional) [3]float32
	// HexColor sets the color for the center vertex of each hex.
	HexColor(h pos.Hex) [3]uint8
	// PointColor sets the color for a hex point vertex that
	// is shared by the hexes h, h.Direction(direction), and h.Direction(direction+1).
	//
	// When using EncodeDetailed:
	// These hex points are shared by the three hexes, so if you
	// return different results for what is actually the same vertex,
	// you will get different results each time you run the encoder.
	//
	// You can use this in a fancy shader to draw borders around
	// hexes if you make it different from HexColor().
	PointColor(h pos.Hex, direction int) [3]uint8
	// EdgeColor returns the color of the rectangle that sits between
	// h and h.Neigbor(direction). This is not used by EncodeDetailed,
	// since that rectangle doesn't exist.
	EdgeColor(h pos.Hex, direction int) [3]uint8
}
