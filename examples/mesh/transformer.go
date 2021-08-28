package mesh

import "github.com/erinpentecost/hex"

// Transformer converts 2-dimensional cartesian space into three dimensions.
// Use this to select which dimension is "up" and do stretching if needed.
type Transformer interface {
	// ConvertTo3D converts some hex vector to 3D cartesian space.
	//
	// glTF defines +Y as up, +Z as forward, and -X as right.
	ConvertTo3D(h hex.Hex, actual hex.HexFractional) [3]float32
	// HexColor sets the color for the center vertex of each hex.
	HexColor(h hex.Hex) [3]uint8
	// PointColor sets the color for a hex point vertex that
	// is shared by the hexes h, h.Direction(direction), and h.Direction(direction+1).
	//
	// You can use this in a fancy shader to draw borders around
	// hexes if you make it different from HexColor().
	PointColor(h hex.Hex, direction int) [3]uint8
	// EdgeColor returns the color of the rectangle that sits between
	// h and h.Neigbor(direction). This is not used by EncodeDetailed,
	// since that rectangle doesn't exist.
	EdgeColor(h hex.Hex, direction int) (top, bottom [3]uint8)
	// PaintEdge should return true if you want to draw the vertical edge between hexes
	// of different heights.
	PaintEdge(h hex.Hex, direction int) bool
	// EdgeSlopeStrength determines if hexes of different heights are shear (0.0) or
	// VERY sloped (1.0).
	EdgeSlopeStrength() float32
}
