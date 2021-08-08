package mesh

import (
	"errors"
	"sort"

	"github.com/erinpentecost/hexcoord/csg"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

// TODO: the detailed encoder makes a BUNCH of extra vertices and triangles.
// it basically is meant to be able to mimic https://en.wikipedia.org/wiki/Giant%27s_Causeway
// each hex gets its own set of vertices, and there's a typically vertical rectangle along
// each side of a hex instead of hexes sharing sides directly.

// There's a potential for a lot of degenerate triangles in this output,
// so dedupe em somehow.

// step 1: assemble each individual hex. there's no vertex sharing here, so it should be easy.
// step 2: for each pair of neighboring hexes, create a rectangle between their shared points.
//		   if this rectangle's area is near zero, don't add the rectangle; just snap the hexes together.
//         else, add the rectangle.

// bonuses:
// 1. pass in the barycentric center of the triangle to the ConvertTo3D() function
//    and use that in combination with the face normals to find better vertex normals

type pointCollection struct {
	hexMap  map[pos.Hex]*hexPoints
	verts   [][3]float32
	normals [][3]float32
	colors  [][3]uint8

	indices []uint16

	transformer Transformer
	area        *csg.Area
}

func newPointCollection(t Transformer, a *csg.Area) *pointCollection {
	if t == nil {
		t = &BaseTransform{}
	}

	b := &pointCollection{
		hexMap:  make(map[pos.Hex]*hexPoints),
		verts:   make([][3]float32, 0),
		normals: make([][3]float32, 0),
		colors:  make([][3]uint8, 0),

		indices: make([]uint16, 0),

		transformer: t,
		area:        a,
	}

	return b
}

func (pc *pointCollection) addHex(h pos.Hex, t Transformer) *hexPoints {
	if hp, ok := pc.hexMap[h]; ok {
		return hp
	}

	hp := newHexPoints(h, t)

	// add indices
	for _, p := range append(hp.points[:], hp.center) {
		idx := uint16(len(pc.verts))
		p.index = idx
		pc.verts = append(pc.verts, p.vert)
		pc.normals = append(pc.normals, p.normal)
		pc.colors = append(pc.colors, p.color)
	}

	// add triangles!
	for i := 0; i < 6; i++ {
		pc.indices = append(pc.indices, hp.center.index, hp.points[i].index, hp.points[pos.BoundFacing(i+1)].index)
	}

	pc.hexMap[h] = hp

	return hp
}

type hexPoints struct {
	h      pos.Hex
	center *point
	points [6]*point
}

type point struct {
	index  uint16
	vert   [3]float32
	normal [3]float32
	color  [3]uint8
}

func newHexPoints(h pos.Hex, t Transformer) *hexPoints {
	hp := &hexPoints{
		h: h,
		center: &point{
			vert:   t.ConvertTo3D(&h, h.ToHexFractional()),
			normal: [3]float32{0, 1, 0},
			color:  t.HexColor(h)},
	}
	for i := 0; i < 6; i++ {
		hp.points[i] = &point{
			vert:   t.ConvertTo3D(&h, pos.Center(h, h.Neighbor(i), h.Neighbor(i+1))),
			color:  t.PointColor(h, pos.BoundFacing(i)),
			normal: [3]float32{0, 1, 0},
		}
	}
	return hp
}

// EncodeDetailedMesh
func EncodeDetailedMesh(a *csg.Area, t Transformer) (doc *gltf.Document, err error) {
	if a.Size() == 0 {
		err = errors.New("can't convert empty area into a mesh")
		return
	}

	if t == nil {
		t = &BaseTransform{}
	}

	doc = gltf.NewDocument()

	hexPoints := newPointCollection(t, a)

	// first pass: add all hexes
	areaHexes := a.Slice()
	for _, a := range areaHexes {
		hexPoints.addHex(a, t)
	}

	// second pass: add edge rects by looking at every neighboring hex pair
	seen := struct{}{}
	seenHexPairs := make(map[[2]pos.Hex]struct{})
	for _, h := range hexPoints.hexMap {
		for i, n := range h.h.Neighbors() {
			// only look at each hex pair once
			keySlice := []pos.Hex{h.h, n}
			sort.Sort(pos.Sort(keySlice))
			key := [2]pos.Hex{keySlice[0], keySlice[1]}
			if _, ok := seenHexPairs[key]; ok {
				continue
			}
			seenHexPairs[key] = seen

			nh, ok := hexPoints.hexMap[n]
			if !ok {
				continue
			}
			// process the pair
			a := h.points[pos.BoundFacing(i)]
			b := h.points[pos.BoundFacing(i-1)]
			c := nh.points[reverseDirection(i-1)]
			d := nh.points[reverseDirection(i)]
			// find area
			rectArea := rectArea(a.vert, b.vert, c.vert)
			if rectArea < 0.01 {
				// snap together degenerate sides
				hexPoints.verts[a.index] = hexPoints.verts[c.index]
				hexPoints.verts[b.index] = hexPoints.verts[d.index]
				continue
			}
			// add rect.
			// TODO: duplicate verts so they aren't shared,
			// and apply correct color
			hexPoints.indices = append(hexPoints.indices, a.index, b.index, c.index)
			hexPoints.indices = append(hexPoints.indices, b.index, d.index, c.index)

		}
	}

	doc.Meshes = []*gltf.Mesh{
		{
			Name: "hexarea",
			Primitives: []*gltf.Primitive{{
				Indices: gltf.Index(modeler.WriteIndices(doc, hexPoints.indices)),
				Attributes: map[string]uint32{
					gltf.POSITION: modeler.WritePosition(doc, hexPoints.verts),
					gltf.COLOR_0:  modeler.WriteColor(doc, hexPoints.colors),
					gltf.NORMAL:   modeler.WriteNormal(doc, hexPoints.normals),
				},
			}},
		},
	}

	doc.Nodes = []*gltf.Node{{Name: "hex", Mesh: gltf.Index(0)}}

	doc.Scene = gltf.Index(0)
	doc.Scenes = []*gltf.Scene{{Name: "hex export", Nodes: []uint32{*doc.Scene}}}

	return doc, nil
}

func rectArea(a, b, c [3]float32) float64 {

	s1 := [3]float32{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
	s2 := [3]float32{b[0] - c[0], b[1] - c[1], b[2] - c[2]}

	return vecMagnitude(vecCross(s1, s2))
}

func reverseDirection(direction int) int {
	return pos.BoundFacing(direction - 3)
}
