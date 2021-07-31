package csg

import (
	"fmt"
	"testing"

	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAreaEqual(t *testing.T) {
	area1 := BigHex(pos.Origin(), 2).Build()
	area2 := BigHex(pos.Origin(), 2).Build()
	area3 := BigHex(pos.Hex{Q: 1, R: 0}, 2).Build()
	area4 := BigHex(pos.Origin(), 2).Subtract(NewArea(pos.Hex{Q: 1, R: 1})).Build()

	assert.True(t, area1.Equal(area2))
	assert.False(t, area1.Equal(area3))
	assert.False(t, area1.Equal(area4))
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

type boundTest struct {
	a        *Area
	b        *Area
	expected Bounding
}

func (b boundTest) assertBound(t *testing.T, name string) {
	t.Helper()
	t.Run(name, func(t *testing.T) {

		// test quick check
		if b.expected == Undefined {
			require.False(t, b.a.mightOverlap(b.b))
			require.False(t, b.b.mightOverlap(b.a))
		} else if b.expected != Distinct {
			require.True(t, b.a.mightOverlap(b.b))
			require.True(t, b.b.mightOverlap(b.a))
		}

		// test actual
		assert.Equal(t, b.expected, b.a.CheckBounding(b.b), "\na=%s\nb=%s", b.a.String(), b.b.String())

		// test reverse
		switch b.expected {
		case Contains:
			assert.Equal(t, ContainedBy, b.b.CheckBounding(b.a), "\na=%s\nb=%s", b.a.String(), b.b.String())
		case ContainedBy:
			assert.Equal(t, Contains, b.b.CheckBounding(b.a), "\na=%s\nb=%s", b.a.String(), b.b.String())
		default:
			assert.Equal(t, b.expected, b.b.CheckBounding(b.a), "\na=%s\nb=%s", b.a.String(), b.b.String())
		}
	})
}

func TestBounding(t *testing.T) {
	tests := []boundTest{
		{a: NewArea(pos.Origin()), b: NewArea(pos.Origin()), expected: Equals},
		{a: NewArea(), b: NewArea(pos.Origin()), expected: Undefined},
		{a: BigHex(pos.Origin(), 4), b: NewArea(pos.Origin()), expected: Contains},
		{a: BigHex(pos.Origin(), 4), b: NewArea(pos.Hex{Q: 100, R: 100}), expected: Distinct},
		{a: BigHex(pos.Origin(), 4), b: BigHex(pos.Hex{Q: 1, R: 1}, 4), expected: Overlap},
		{a: BigHex(pos.Origin(), 5), b: BigHex(pos.Origin(), 5).Subtract(NewArea(pos.Hex{Q: 1, R: 1})).Build(), expected: Contains},
		{a: Rectangle(pos.Hex{Q: 5, R: 5}, pos.Hex{Q: 10, R: 10}).Union(NewArea(pos.Origin())).Build(), b: NewArea(pos.Origin()), expected: Contains},
	}
	for i, test := range tests {
		test.assertBound(t, fmt.Sprintf("%d", i))
	}
}
