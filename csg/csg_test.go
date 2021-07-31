package csg

import (
	"testing"

	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
)

func TestRingVsBigHex(t *testing.T) {
	t.Run("subtract", func(t *testing.T) {
		for i := 0; i <= 5; i++ {

			area1 := Ring(pos.Origin(), i).Build()
			area2 := BigHex(pos.Origin(), i).Subtract(BigHex(pos.Origin(), i-1)).Build()

			assert.True(t, area1.Equal(area2), "Areas are not equal.")
		}
	})
	t.Run("union", func(t *testing.T) {
		for i := 0; i <= 5; i++ {
			ringBuilder := NewBuilder()
			for r := 0; r <= i; r++ {
				ringBuilder = ringBuilder.Union(Ring(pos.Origin(), r))
			}

			area1 := ringBuilder.Build()
			area2 := BigHex(pos.Origin(), i)

			assert.True(t, area1.Equal(area2), "Areas are not equal.")
		}
	})
}

func TestAreaEqual(t *testing.T) {
	area1 := Ring(pos.Origin(), 1).Build()
	area2 := Ring(pos.Origin(), 1).Build()
	area3 := Ring(pos.Origin(), 2).Build()

	assert.True(t, area1.Equal(area2), "Areas are not equal.")
	assert.False(t, area1.Equal(area3), "Areas are equal.")
}

func TestAreaIntersection(t *testing.T) {

	identity := BigHex(pos.Origin(), 10).Intersection(BigHex(pos.Origin(), 10)).Build()

	assert.True(t,
		identity.Equal(BigHex(pos.Origin(), 10)),
		"Intersection failed on matched input.")

	ringCheck := Ring(pos.Origin(), 4).Intersection(BigHex(pos.Origin(), 10)).Build()

	assert.True(t,
		ringCheck.Equal(Ring(pos.Origin(), 4)),
		"Intersection failed with unmatched input.")
}

func TestTriangle(t *testing.T) {
	// points for a big triangle
	points := []pos.Hex{
		{Q: 1, R: -2},
		{Q: 1, R: 1},
		{Q: -2, R: 1},
	}
	outline := Line(append(points, points[0])...)
	expectedOutline := []pos.Hex{
		{Q: 1, R: -2},
		{Q: 1, R: -1},
		{Q: 1, R: 0},
		{Q: 1, R: 1},
		{Q: 0, R: 1},
		{Q: -1, R: 1},
		{Q: -2, R: 1},
		{Q: -1, R: 0},
		{Q: 0, R: -1},
	}
	expectedOutlineArea := NewArea(expectedOutline...)
	assert.True(t, expectedOutlineArea.Equal(outline), "expected=%s\nactual=  %s\n", expectedOutlineArea.String(), outline.String())

	fill := Polygon(points...).Build()
	expectedFillArea := NewArea(append(expectedOutline, pos.Origin())...)
	assert.True(t, expectedFillArea.Equal(fill), "expected=%s\nactual=  %s\n", expectedFillArea.String(), fill.String())
}
