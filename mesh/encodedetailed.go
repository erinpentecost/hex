package mesh

import (
	"errors"
	"sort"

	"github.com/erinpentecost/hexcoord/csg"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

// TODO:
// 1. pass in the barycentric center of each triangle to the ConvertTo3D() function
//    and use that in combination with the face normals to find smooth vertex normals

type pointCollection struct {
	hexMap map[pos.Hex]*hexPoints
	verts  [][3]float32
	colors [][3]uint8

	indices []uint16

	transformer Transformer
	area        *csg.Area
}

func newPointCollection(t Transformer, a *csg.Area) *pointCollection {
	if t == nil {
		t = &BaseTransform{}
	}

	b := &pointCollection{
		hexMap: make(map[pos.Hex]*hexPoints),
		verts:  make([][3]float32, 0),
		colors: make([][3]uint8, 0),

		indices: make([]uint16, 0),

		transformer: t,
		area:        a,
	}

	return b
}

func (pc *pointCollection) addOrGetHex(h pos.Hex, t Transformer, invisible bool) *hexPoints {
	if hp, ok := pc.hexMap[h]; ok {
		return hp
	}

	hp := &hexPoints{
		invisible: invisible,
		h:         h,
		center: &point{
			vert:   t.ConvertTo3D(h, h.ToHexFractional()),
			normal: [3]float32{0, 1, 0},
			color:  t.HexColor(h)},
	}
	for i := 0; i < 6; i++ {
		hp.points[i] = &point{
			vert:   t.ConvertTo3D(h, pos.Center(h, h.Neighbor(i), h.Neighbor(i+1))),
			color:  t.PointColor(h, pos.BoundFacing(i)),
			normal: [3]float32{0, 1, 0},
		}
	}

	pc.hexMap[h] = hp

	if invisible {
		return hp
	}

	// add indices for hex top
	for _, p := range append(hp.points[:], hp.center) {
		idx := uint16(len(pc.verts))
		p.index = idx
		pc.verts = append(pc.verts, p.vert)
		pc.colors = append(pc.colors, p.color)
	}

	// add triangles for hex top
	for i := 0; i < 6; i++ {
		pc.indices = append(pc.indices, hp.center.index, hp.points[i].index, hp.points[pos.BoundFacing(i+1)].index)
	}

	return hp
}

type hexPoints struct {
	h         pos.Hex
	invisible bool
	center    *point
	points    [6]*point
}

type point struct {
	// index is the index for the hex face index
	index uint16
	// rectIndex is the index for the vertical side face, if present.
	// if not present, this is 0.
	rectIndex uint16
	vert      [3]float32
	normal    [3]float32
	color     [3]uint8
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

	// first pass: add all hexes in the area
	areaHexes := a.Slice()
	for _, a := range areaHexes {
		hexPoints.addOrGetHex(a, t, false)
	}

	// second pass: add edge rects by looking at every neighboring hex pair
	seen := struct{}{}
	seenHexPairs := make(map[[2]pos.Hex]struct{})
	for _, h := range hexPoints.hexMap {
		if h.invisible {
			continue
		}

		for i, n := range h.h.Neighbors() {
			// only look at each hex pair once
			keySlice := []pos.Hex{h.h, n}
			sort.Sort(pos.Sort(keySlice))
			key := [2]pos.Hex{keySlice[0], keySlice[1]}
			if _, ok := seenHexPairs[key]; ok {
				continue
			}
			seenHexPairs[key] = seen

			// get points for the rect
			a := h.points[pos.BoundFacing(i)]
			b := h.points[pos.BoundFacing(i-1)]
			var c, d *point

			nh := hexPoints.addOrGetHex(n, t, true)

			c = nh.points[reverseDirection(i-1)]
			d = nh.points[reverseDirection(i)]

			// don't draw rects unless this hex is taller on the border
			if nh.invisible && a.vert[1] < c.vert[1] {
				continue
			}

			// find area
			rectArea := rectArea(a.vert, b.vert, c.vert)
			if rectArea < 0.01 && !nh.invisible {
				// snap together degenerate sides
				a.vert = c.vert
				b.vert = d.vert
				hexPoints.verts[a.index] = hexPoints.verts[c.index]
				hexPoints.verts[b.index] = hexPoints.verts[d.index]
				continue
			}

			// add rect.
			start := uint16(len(hexPoints.verts))
			hexPoints.verts = append(hexPoints.verts, a.vert, b.vert, c.vert, d.vert)

			if a.vert[1] > c.vert[1] {
				topColor, bottomColor := t.EdgeColor(h.h, i)
				hexPoints.colors = append(hexPoints.colors, topColor, topColor, bottomColor, bottomColor)
			} else {
				topColor, bottomColor := t.EdgeColor(nh.h, i)
				hexPoints.colors = append(hexPoints.colors, bottomColor, bottomColor, topColor, topColor)
			}

			hexPoints.indices = append(hexPoints.indices, start, start+1, start+2)
			hexPoints.indices = append(hexPoints.indices, start+1, start+3, start+2)

			a.rectIndex = start
			b.rectIndex = start + 1
			c.rectIndex = start + 2
			d.rectIndex = start + 3
		}
	}

	// third pass: add triangle between each triple
	seenHexTriples := make(map[[3]pos.Hex]struct{})
	for _, h := range hexPoints.hexMap {
		if h.invisible {
			continue
		}

		for i, n1 := range h.h.Neighbors() {
			n2 := h.h.Neighbor(i + 1)
			// only look at each hex triple once
			keySlice := []pos.Hex{h.h, n1, n2}
			sort.Sort(pos.Sort(keySlice))
			key := [3]pos.Hex{keySlice[0], keySlice[1], keySlice[2]}
			if _, ok := seenHexTriples[key]; ok {
				continue
			}
			seenHexTriples[key] = seen

			// get points for the triangle
			a := h.points[pos.BoundFacing(i)]
			var b, c *point

			nh1 := hexPoints.addOrGetHex(n1, t, false)
			b = nh1.points[reverseDirection(i-1)]

			nh2 := hexPoints.addOrGetHex(n2, t, false)
			c = nh2.points[reverseDirection(i+1)] //-1

			// don't draw triangle on border unless this hex is taller
			// TODO: I don't like this. potential for missing triangles.
			if nh1.invisible && nh2.invisible && a.vert[1] < c.vert[1] {
				continue
			}

			// don't bother if there are no rectindexes
			if a.rectIndex == 0 || b.rectIndex == 0 || c.rectIndex == 0 {
				continue
			}

			// find verts used by the rects and attach em
			hexPoints.indices = append(hexPoints.indices, a.rectIndex, b.rectIndex, c.rectIndex)

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
