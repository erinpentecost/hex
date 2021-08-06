package mesh

import (
	"testing"

	"github.com/erinpentecost/hexcoord/csg"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/qmuntal/gltf"
	"github.com/stretchr/testify/require"
)

func TestDrawArea(t *testing.T) {
	area := csg.BigHex(pos.Origin(), 3).Subtract(csg.Line(pos.Hex{Q: -2, R: 0}, pos.Hex{Q: 2, R: 0})).Build()
	doc, err := EncodeOptimizedMesh(area, nil)
	require.NoError(t, err)
	gltf.SaveBinary(doc, "sample.glb")
}
