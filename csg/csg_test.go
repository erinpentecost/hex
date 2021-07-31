package csg

import (
	"testing"

	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
)

func TestAreaEqual(t *testing.T) {
	area1 := BigHex(pos.Origin(), 1).Build()
	area2 := BigHex(pos.Origin(), 1).Build()
	area3 := BigHex(pos.Origin(), 2).Build()

	assert.True(t, area1.Equal(area2), "Areas are not equal.")
	assert.False(t, area1.Equal(area3), "Areas are equal.")
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
