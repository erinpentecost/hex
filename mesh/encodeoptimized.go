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
	// EmbedNormals should be true if you want edge information embedded in the NORMAL
	// attributes.
	EmbedNormals() bool
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

	// this is positionindex to some vec
	indexToNormal map[uint16]([3]float32)

	transformer OptimizedTransformer
}

func newBufferBuilder(t OptimizedTransformer) *bufferBuilder {
	if t == nil {
		t = &BaseTransform{}
	}
	if t.EmbedNormals() {
		panic("not implemented yet")
	}

	b := &bufferBuilder{
		verts:         make([][3]float32, 0),
		indices:       make([]uint16, 0),
		hexesToIndex:  make(map[[3]pos.Hex]uint16),
		indexToNormal: make(map[uint16][3]float32),
		transformer:   t,
	}

	/*
		// find normal
		o1 := t.ConvertToOptimized3D(pos.OriginFractional())
		o2 := t.ConvertToOptimized3D(pos.HexFractional{Q: 0.0, R: 100.0})
		o3 := t.ConvertToOptimized3D(pos.HexFractional{Q: 100.0, R: 100.0})
		e1 := vecSub(o1, o2)
		e2 := vecSub(o1, o3)
		norm := vecCross(e1, e2)
		normMag := vecMagnitude(norm)
		norm[0] = norm[0] / float32(normMag)
		norm[1] = norm[1] / float32(normMag)
		norm[2] = norm[2] / float32(normMag)
		b.unitNormal = norm
		b.unitNormalIdx = uint16(len(b.verts))
		b.verts = append(b.verts, b.unitNormal)

		// calc unit vecs that point in directions to edges
		var ev [6][3]float32
		if t.EmbedNormals() {
			ev = normalVecs(t)
			// add other handy vecs
			for i, edgeVert := range ev {
				newVert := len(b.verts)
				b.verts = append(b.verts, edgeVert)
				b.unitEdgeVecIdx[i] = uint16(newVert)
			}
		} else {
			// just use unit normal for everything
			for i := range ev {
				ev[i] = b.unitNormal
				b.unitEdgeVecIdx[i] = b.unitNormalIdx
			}
		}*/

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
	b.hexesToIndex[h] = uint16(newVert)

	return uint16(newVert)
}

// normalVecs returns a direction-to-edge-indexed array of
// unit vectors pointing in those directions.
func normalVecs(t OptimizedTransformer) (norms [6][3]float32) {
	o := pos.Origin()
	for i := 0; i < 6; i++ {
		norms[i] = t.ConvertToOptimized3D(pos.Center(o.Neighbor(i), o.Neighbor(i+1), o))
		mag := vecMagnitude(norms[i])
		norms[i][0] = norms[i][0] / float32(mag)
		norms[i][1] = norms[i][2] / float32(mag)
		norms[i][1] = norms[i][2] / float32(mag)
	}
	return norms
}

func (b *bufferBuilder) addNewHex(h pos.Hex) {
	// we need the internal vertex
	originIndex := b.getIndexFromHexes([3]pos.Hex{h, h, h})
	neighborIndexes := [6]uint16{}
	// now we need to get all the neighbor vertices
	for i := 0; i < 6; i++ {
		neighborIndexes[i] = b.getIndexFromHexes([3]pos.Hex{h.Neighbor(i), h.Neighbor(i + 1), h})
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

	bb := newBufferBuilder(t)

	for _, h := range a.Slice() {
		bb.addNewHex(h)
	}

	attrs, _ := modeler.WriteAttributesInterleaved(doc, modeler.Attributes{
		Position: bb.verts,
	})

	doc.Meshes = []*gltf.Mesh{
		{
			Name: "hexarea",
			Primitives: []*gltf.Primitive{{
				Indices:    gltf.Index(modeler.WriteIndices(doc, bb.indices)),
				Attributes: attrs,
			}},
		},
	}

	doc.Nodes = []*gltf.Node{{Name: "hex", Mesh: gltf.Index(0)}}

	doc.Scene = gltf.Index(0)
	doc.Scenes = []*gltf.Scene{{Name: "hex export", Nodes: []uint32{*doc.Scene}}}

	return doc, nil
}
