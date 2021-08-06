package mesh

import (
	"errors"
	"sort"

	"github.com/erinpentecost/hexcoord/csg"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

// OptimizedTransformer converts 2-dimensional cartesian space into three dimensions.
// Use this to select which dimension is "up" and do stretching if needed.
type OptimizedTransformer interface {
	// ConvertToOptimized3D converts some hex vector to 3D cartesian space.
	// glTF defines +Y as up, +Z as forward, and -X as right.
	ConvertToOptimized3D(actual pos.HexFractional) [3]float32

	HexColor(h pos.Hex) [3]uint8

	EdgeColor(h, n1, n2 pos.Hex) [3]uint8
}

type bufferBuilder struct {
	// verts are actual vertices that exist in 3d space
	verts [][3]float32
	// indices are a list of indexes into the verts array.
	// the length of this should always be a multiple of 3,
	// since they describe the points of triangles in the mesh.
	indices []uint16

	// hexesToIndex maps the center point of three neighboring hexes
	// to some index in the indices field.
	// this is used for deduping.
	hexesToIndex map[[3]pos.Hex]uint16

	colors [][3]uint8

	transformer OptimizedTransformer

	area *csg.Area
}

func newBufferBuilder(t OptimizedTransformer, a *csg.Area) *bufferBuilder {
	if t == nil {
		t = &BaseTransform{}
	}

	b := &bufferBuilder{
		verts:        make([][3]float32, 0),
		indices:      make([]uint16, 0),
		hexesToIndex: make(map[[3]pos.Hex]uint16),
		colors:       make([][3]uint8, 0),
		transformer:  t,
		area:         a,
	}

	return b
}

// gets the index from the hex edge point.
// this does not add the index to the indices array
func (b *bufferBuilder) getIndexFromHexes(h [3]pos.Hex) uint16 {
	// this is so bad. stick in a consistent order
	vertTriple := h[:]
	sort.Sort(pos.Sort(vertTriple))
	h = [3]pos.Hex{vertTriple[0], vertTriple[1], vertTriple[2]}

	// then do the lookup
	if found, ok := b.hexesToIndex[h]; ok {
		return uint16(found)
	}

	newVert := len(b.verts)
	b.verts = append(b.verts, b.transformer.ConvertToOptimized3D(pos.Center(h[:]...)))
	b.colors = append(b.colors, [3]uint8{})
	b.hexesToIndex[h] = uint16(newVert)

	return uint16(newVert)
}

func (b *bufferBuilder) addNewHex(h pos.Hex) {
	// we need the internal vertex
	originIndex := b.getIndexFromHexes([3]pos.Hex{h, h, h})
	b.colors[originIndex] = b.transformer.HexColor(h)

	neighborIndexes := [6]uint16{}

	// now we need to get all the neighbor vertices
	for i := 0; i < 6; i++ {
		n1 := h.Neighbor(i)
		n2 := h.Neighbor(i + 1)
		neighborIndexes[i] = b.getIndexFromHexes([3]pos.Hex{n1, n2, h})

		b.colors[neighborIndexes[i]] = b.transformer.EdgeColor(h, n1, n2)
	}

	// add a bunch of triangles now
	for i := 0; i < 6; i++ {
		b.indices = append(b.indices, originIndex, neighborIndexes[i], neighborIndexes[pos.BoundFacing(i+1)])

	}
}

// EncodeOptimizedMesh is suitable for when you need a gigantic resolution mesh.
// It spits out a simple flat mesh with deduped vertices.
// The normals of the vertices encode information about what type of vertex it is.
//
// NORMAL encoding:
// 1. be normal, but point downward for, hex center vertices.
// 2. be normal to the hex face for shared internal vertices
// 3. point away from the hex area for Concave boundary vertices
// 4. point toward the hex area for Convex boundary verticesshared by 3 hexes.
func EncodeOptimizedMesh(a *csg.Area, t OptimizedTransformer) (doc *gltf.Document, err error) {
	if a.Size() == 0 {
		err = errors.New("can't convert empty area into a mesh")
		return
	}

	doc = gltf.NewDocument()

	bb := newBufferBuilder(t, a)

	for _, h := range a.Slice() {
		bb.addNewHex(h)
	}

	doc.Meshes = []*gltf.Mesh{
		{
			Name: "hexarea",
			Primitives: []*gltf.Primitive{{
				Indices: gltf.Index(modeler.WriteIndices(doc, bb.indices)),
				Attributes: map[string]uint32{
					gltf.POSITION: modeler.WritePosition(doc, bb.verts),
					gltf.COLOR_0:  modeler.WriteColor(doc, bb.colors),
				},
			}},
		},
	}

	doc.Nodes = []*gltf.Node{{Name: "hex", Mesh: gltf.Index(0)}}

	doc.Scene = gltf.Index(0)
	doc.Scenes = []*gltf.Scene{{Name: "hex export", Nodes: []uint32{*doc.Scene}}}

	return doc, nil
}
