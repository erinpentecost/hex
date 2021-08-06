package mesh

import (
	"github.com/erinpentecost/hexcoord/csg"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/qmuntal/gltf"
)

type DetailedTransformer interface {
	ConvertToDetailed3D(hd pos.Hex, actual pos.HexFractional) [3]float32
}

// TODO: the detailed encoder makes a BUNCH of extra vertices and triangles.
// it basically is meant to be able to mimic https://en.wikipedia.org/wiki/Giant%27s_Causeway
// each hex gets its own set of vertices, and there's a typically vertical rectangle along
// each side of a hex instead of hexes sharing sides directly.

// EncodeDetailedMesh
func EncodeDetailedMesh(a *csg.Area, t DetailedTransformer) (doc *gltf.Document, err error) {
	panic("not implemented")
}
