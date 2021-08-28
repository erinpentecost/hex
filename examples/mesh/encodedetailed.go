package mesh

import (
	"errors"
	"math"
	"sort"

	"github.com/erinpentecost/hex"
	"github.com/erinpentecost/hex/area"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
	"github.com/ungerik/go3d/vec3"
)

var seen struct{}

// TODO:
// 1. pass in the barycentric center of each triangle to the ConvertTo3D() function
//    and use that in combination with the face normals to find smooth vertex normals

type pointCollection struct {
	hexMap  map[hex.Hex]*hexPointCollection
	verts   [][3]float32
	colors  [][3]uint8
	normals [][3]float32

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
		pc.normals = append(pc.normals, [3]float32{0.0, 1.0, 0.0})
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
	vert      vec3.T
	normal    vec3.T
	color     [3]uint8
}

func pushToCenter(hexPoints *pointCollection, center *vec3.T, toMove *point, t float32) {
	center[1] = toMove.vert[1]
	toMove.vert = vec3.Interpolate(copy(toMove.vert), center, t)
	hexPoints.verts[toMove.index] = toMove.vert
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

	// second pass: look at each hex triple and identify if the shared point
	// has 3 levels. if it does, find the middle point and snap it to line
	// formed by the higher and lower point.
	slopeStrength := float32(math.Max(0.0, math.Min(1.0, float64(t.EdgeSlopeStrength())))) * 0.2
	seenHexTriples := make(map[[3]hex.Hex]struct{})
	for _, hp := range areaHexes {
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
			var upperHx *hexPointCollection
			var upper, mid, lower *point

			upperHx = h
			upper = h.points[hex.BoundFacing(i)]
			midHx := hexPoints.addOrGetHex(n1, t, true)
			mid = midHx.points[reverseDirection(i-1)]
			lowerHx := hexPoints.addOrGetHex(n2, t, true)
			lower = lowerHx.points[reverseDirection(i+1)]

			if upper.vert[1] < mid.vert[1] {
				upper, mid = mid, upper
				upperHx, midHx = midHx, upperHx
			}
			if mid.vert[1] < lower.vert[1] {
				mid, lower = lower, mid
				midHx, lowerHx = lowerHx, midHx
			}
			if upper.vert[1] < mid.vert[1] {
				upper, mid = mid, upper
				upperHx, midHx = midHx, upperHx
			}

			if upper.vert[1]+0.0001 > lower.vert[1] && lower.vert[1]+0.0001 > upper.vert[1] {
				// they are flat, so no snapping
				continue // TODO: if I don't do this then everything REALLY breaks
			}

			// push corners away from eachother. this makes overhangs less likely
			upperCenterVec := upperHx.center.vert
			midCenterVec := midHx.center.vert
			lowerCenterVec := lowerHx.center.vert
			heightDiff := (upperCenterVec[1] - lowerCenterVec[1]) / (midCenterVec[1] - lowerCenterVec[1])
			//heightDiff = float32(math.Log(float64(heightDiff) + math.SqrtE))
			heightDiff = float32(math.Log(float64(heightDiff)))
			if heightDiff > 1.0 {
				heightDiff = 1.0
			}

			pushTargetVec := vec3.Interpolate(&midCenterVec, &upperCenterVec, heightDiff)

			if !upperHx.invisible {
				pushToCenter(hexPoints, &pushTargetVec, upper, slopeStrength)
			}
			if !lowerHx.invisible {
				pushToCenter(hexPoints, &pushTargetVec, lower, -1*slopeStrength)
			}

			// snap mid to point along line formed by upper and lower
			if midHx.invisible {
				continue
			}
			t := (mid.vert[1] - lower.vert[1]) / (upper.vert[1] - lower.vert[1])
			newMid := vec3.Interpolate((*vec3.T)(&lower.vert), (*vec3.T)(&upper.vert), t)

			mid.vert = newMid
			hexPoints.verts[mid.index] = mid.vert
		}
	}

	// third pass: add edge rects by looking at every neighboring hex pair
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

			if !t.PaintEdge(h.h, i) {
				continue
			}

			// get points for the rect
			a := h.points[hex.BoundFacing(i)]
			b := h.points[hex.BoundFacing(i-1)]
			var c, d *point

			nh := hexPoints.addOrGetHex(n, t, true)

			c = nh.points[reverseDirection(i-1)]
			d = nh.points[reverseDirection(i)]

			// find normals, which should be average of these two triangles
			v1 := vec3.Cross(copy(a.vert).Sub(&b.vert), copy(c.vert).Sub(&b.vert))
			v1 = *v1.Normalize()
			v2 := vec3.Cross(copy(b.vert).Sub(&d.vert), copy(c.vert).Sub(&d.vert))
			v2 = *v2.Normalize()
			shared := vec3.Interpolate(&v1, &v2, 0.5)
			shared = *shared.Normalize()

			// normals will be 0 if the rect is degenerate
			if v1.IsZero() || v2.IsZero() || shared.IsZero() {
				continue
			}

			// soften up originals a bit
			v1 = vec3.Interpolate(&v1, &shared, 0.5)
			v1 = *v1.Normalize()
			v2 = vec3.Interpolate(&v2, &shared, 0.5)
			v2 = *v2.Normalize()

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

			// add normals
			hexPoints.normals = append(hexPoints.normals, v1)
			hexPoints.normals = append(hexPoints.normals, shared)
			hexPoints.normals = append(hexPoints.normals, shared)
			hexPoints.normals = append(hexPoints.normals, v2)

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

func reverseDirection(direction int) int {
	return hex.BoundFacing(direction - 3)
}

func copy(vec vec3.T) *vec3.T {
	return &vec3.T{vec[0], vec[1], vec[2]}
}
