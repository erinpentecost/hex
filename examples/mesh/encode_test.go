package mesh

import (
	"testing"

	"github.com/erinpentecost/hex"
	"github.com/erinpentecost/hex/area"
	"github.com/qmuntal/gltf"
	"github.com/stretchr/testify/require"
)

func TestDrawArea(t *testing.T) {
	area := area.BigHex(hex.Origin(), 3).
		Subtract(area.Line(hex.Hex{Q: -2, R: 0}, hex.Hex{Q: 2, R: 0})).
		Subtract(area.BigHex(hex.Hex{Q: 2, R: 1}, 2)).
		Union(area.NewArea(hex.Hex{Q: 1, R: 2})).
		Build()

	doc, err := EncodeDetailedMesh(area, &BaseTransform{area: *area})
	require.NoError(t, err)
	gltf.SaveBinary(doc, "detail_sample.glb")
}

func TestDir(t *testing.T) {
	p := hex.Hex{Q: 2, R: -1}
	for i := -10; i < 10; i++ {
		d := p.Neighbor(i)
		dd := d.Neighbor(reverseDirection(i))
		require.Equal(t, p, dd)
	}
}
