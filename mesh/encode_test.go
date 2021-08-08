package mesh

import (
	"testing"

	"github.com/erinpentecost/hexcoord/csg"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/qmuntal/gltf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDrawArea(t *testing.T) {
	area := csg.BigHex(pos.Origin(), 3).
		Subtract(csg.Line(pos.Hex{Q: -2, R: 0}, pos.Hex{Q: 2, R: 0})).
		Union(csg.Line(pos.Hex{Q: 4, R: 4}, pos.Hex{Q: 4, R: 5})).
		Build()

	doc, err := EncodeOptimizedMesh(area, nil)
	require.NoError(t, err)
	gltf.SaveBinary(doc, "optimized_sample.glb")

	doc, err = EncodeDetailedMesh(area, nil)
	require.NoError(t, err)
	gltf.SaveBinary(doc, "detail_sample.glb")
}

func TestDir(t *testing.T) {
	p := pos.Hex{Q: 2, R: -1}
	for i := -10; i < 10; i++ {
		d := p.Neighbor(i)
		dd := d.Neighbor(reverseDirection(i))
		require.Equal(t, p, dd)
	}
}

func TestRectArea(t *testing.T) {

	check := func(t *testing.T, a, b, c, d [3]float32, e float64) {
		t.Helper()
		area := rectArea(a, b, c)
		assert.Equal(t, e, area)
		area = rectArea(b, c, d)
		assert.Equal(t, e, area)
		area = rectArea(c, d, a)
		assert.Equal(t, e, area)
		area = rectArea(d, a, b)
		assert.Equal(t, e, area)
	}

	a := [3]float32{10, 10, 10}
	b := [3]float32{13, 10, 10}
	c := [3]float32{13, 14, 10}
	d := [3]float32{10, 14, 10}

	check(t, a, b, c, d, float64(12))

	check(t, a, b, a, b, float64(0))
}
