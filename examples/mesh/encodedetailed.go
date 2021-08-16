package mesh

import (
	"errors"
	"fmt"
	"sort"

	"github.com/erinpentecost/hex"
	"github.com/erinpentecost/hex/area"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
	"github.com/ungerik/go3d/vec3"
)

// TODO:
// 1. pass in the barycentric center of each triangle to the ConvertTo3D() function
//    and use that in combination with the face normals to find smooth vertex normals

type pointCollection struct {
	hexMap map[hex.Hex]*hexPointCollection
	verts  [][3]float32
	colors [][3]uint8

	indices []uint16

	transformer Transformer
	area        *area.Area
}

func newPointCollection(t Transformer, a *area.Area) *pointCollection {
	if t == nil {
		t = &BaseTransform{}
	}

	b := &pointCollection{
		hexMap: make(map[hex.Hex]*hexPointCollection),
		verts:  make([][3]float32, 0),
		colors: make([][3]uint8, 0),

		indices: make([]uint16, 0),

		transformer: t,
		area:        a,
	}

	return b
}

func (pc *pointCollection) addOrGetHex(h hex.Hex, t Transformer, invisible bool) *hexPointCollection {
	if hp, ok := pc.hexMap[h]; ok {
		return hp
	}

	hp := &hexPointCollection{
		invisible: invisible,
		h:         h,
		center: &point{
			vert:   t.ConvertTo3D(h, h.ToHexFractional()),
			normal: [3]float32{0, 1, 0},
			color:  t.HexColor(h)},
	}
	for i := 0; i < 6; i++ {
		hp.points[i] = &point{
			vert:   t.ConvertTo3D(h, hex.Center(h, h.Neighbor(i), h.Neighbor(i+1))),
			color:  t.PointColor(h, hex.BoundFacing(i)),
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
	// TODO: calculate normals for each vertex.
	//       get face normal of each pair of triangles, then average them, apply vertex normal
	//       of shared vertex
	for i := 0; i < 6; i++ {
		pc.indices = append(pc.indices, hp.center.index, hp.points[i].index, hp.points[hex.BoundFacing(i+1)].index)
	}

	return hp
}

type hexPointCollection struct {
	h         hex.Hex
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
func EncodeDetailedMesh(a *area.Area, t Transformer) (doc *gltf.Document, err error) {
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

	seen := struct{}{}

	// second pass: snap Y if edges are close enough
	seenHexPairs := make(map[[2]hex.Hex]struct{})
	for _, hp := range areaHexes {
		h := hexPoints.hexMap[hp]

		if h.invisible {
			continue
		}

		for i, n := range h.h.Neighbors() {
			// only look at each hex pair once
			keySlice := []hex.Hex{h.h, n}
			sort.Sort(hex.Sort(keySlice))
			key := [2]hex.Hex{keySlice[0], keySlice[1]}
			if _, ok := seenHexPairs[key]; ok {
				continue
			}
			seenHexPairs[key] = seen

			// get points for the rect
			a := h.points[hex.BoundFacing(i)]
			b := h.points[hex.BoundFacing(i-1)]
			var c, d *point

			nh := hexPoints.addOrGetHex(n, t, true)

			c = nh.points[reverseDirection(i-1)]
			d = nh.points[reverseDirection(i)]

			// find area in the Y plane
			rectArea := rectYArea(a.vert, b.vert, c.vert)
			if rectArea < 0.001 && !nh.invisible {
				// snap together degenerate sides
				a.vert = c.vert
				b.vert = d.vert
				hexPoints.verts[a.index] = hexPoints.verts[c.index]
				hexPoints.verts[b.index] = hexPoints.verts[d.index]
				continue
			}
		}
	}

	// third pass: look at each hex triple and identify if the shared point
	// has 3 levels. if it does, find the middle point and snap it to line
	// formed by the higher and lower point.
	seenHexTriples := make(map[[3]hex.Hex]struct{})
	for di, hp := range areaHexes {
		h := hexPoints.hexMap[hp]
		if h.invisible {
			continue
		}

		for i, n1 := range h.h.Neighbors() {
			n2 := h.h.Neighbor(i + 1)
			// only look at each hex triple once
			keySlice := []hex.Hex{h.h, n1, n2}
			sort.Sort(hex.Sort(keySlice))
			key := [3]hex.Hex{keySlice[0], keySlice[1], keySlice[2]}
			if _, ok := seenHexTriples[key]; ok {
				continue
			}
			seenHexTriples[key] = seen

			// get points for the triangle
			var upper, mid, lower *point
			{
				upperHx := h
				upper = h.points[hex.BoundFacing(i)]
				midHx := hexPoints.addOrGetHex(n1, t, true)
				mid = midHx.points[reverseDirection(i-1)]
				lowerHx := hexPoints.addOrGetHex(n2, t, true)
				lower = lowerHx.points[reverseDirection(i+1)]

				if upper.vert[1] < lower.vert[1] {
					upper, lower = lower, upper
					upperHx, lowerHx = lowerHx, upperHx
				}
				if mid.vert[1] < lower.vert[1] {
					mid, lower = lower, mid
					midHx, lowerHx = lowerHx, midHx
				}
				if upper.vert[1] < lower.vert[1] {
					upper, lower = lower, upper
					upperHx, lowerHx = lowerHx, upperHx
				}

				if midHx.invisible {
					continue
				}
			}

			if upper.vert[1]+0.0001 > lower.vert[1] && lower.vert[1]+0.0001 > upper.vert[1] {
				// they are flat, so no snapping
				continue // TODO: if I don't do this then everything REALLY breaks
			}

			// snap mid to point along line formed by upper and lower
			t := (mid.vert[1] - lower.vert[1]) / (upper.vert[1] - lower.vert[1])
			newMid := vec3.Interpolate((*vec3.T)(&lower.vert), (*vec3.T)(&upper.vert), t)

			if vec3.Distance(&newMid, (*vec3.T)(&mid.vert)) > 0.5 {
				panic(fmt.Sprintf("what? %d/%d", di, len(areaHexes)))
			}

			mid.vert = newMid
			hexPoints.verts[mid.index] = mid.vert
		}
	}

	// fourth pass: add edge rects by looking at every neighboring hex pair
	seenHexPairs = make(map[[2]hex.Hex]struct{})
	for _, hp := range areaHexes {
		h := hexPoints.hexMap[hp]

		if h.invisible {
			continue
		}

		for i, n := range h.h.Neighbors() {
			// only look at each hex pair once
			keySlice := []hex.Hex{h.h, n}
			sort.Sort(hex.Sort(keySlice))
			key := [2]hex.Hex{keySlice[0], keySlice[1]}
			if _, ok := seenHexPairs[key]; ok {
				continue
			}
			seenHexPairs[key] = seen

			// get points for the rect
			a := h.points[hex.BoundFacing(i)]
			b := h.points[hex.BoundFacing(i-1)]
			var c, d *point

			nh := hexPoints.addOrGetHex(n, t, true)

			c = nh.points[reverseDirection(i-1)]
			d = nh.points[reverseDirection(i)]

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

func rectYArea(a, b, c [3]float32) float64 {

	s1 := [3]float32{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
	s2 := [3]float32{b[0] - c[0], a[1] - b[1], b[2] - c[2]}

	cross := vecCross(s1, s2)
	cross[0] = 0
	cross[2] = 0

	return vecMagnitude(cross)
}

func rectArea(a, b, c [3]float32) float64 {

	s1 := [3]float32{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
	s2 := [3]float32{b[0] - c[0], b[1] - c[1], b[2] - c[2]}

	return vecMagnitude(vecCross(s1, s2))
}

func reverseDirection(direction int) int {
	return hex.BoundFacing(direction - 3)
}
