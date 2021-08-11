package area

import (
	"fmt"
	"testing"

	"github.com/erinpentecost/hex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAreaEqual(t *testing.T) {
	area1 := BigHex(hex.Origin(), 2).Build()
	area2 := BigHex(hex.Origin(), 2).Build()
	area3 := BigHex(hex.Hex{Q: 1, R: 0}, 2).Build()
	area4 := BigHex(hex.Origin(), 2).Subtract(NewArea(hex.Hex{Q: 1, R: 1})).Build()

	assert.True(t, area1.Equals(area2))
	assert.False(t, area1.Equals(area3))
	assert.False(t, area1.Equals(area4))
}

func TestIdentity(t *testing.T) {
	orig := BigHex(hex.Origin(), 4).Build()
	translate := orig.Translate(hex.Origin()).Build()
	rotated := orig.Rotate(hex.Origin(), 0).Build()

	assert.True(t, orig.Equals(translate), "expected=%s\nactual=%s", orig.String(), rotated.String())
	assert.True(t, orig.Equals(rotated), "expected=%s\nactual=%s", orig.String(), rotated.String())
}

func TestTranslate(t *testing.T) {
	points := BigHex(hex.Origin(), 4).Build().Slice()
	for _, point := range points {
		for _, offset := range points {
			newPoint := NewArea(point).Translate(offset).Build()
			expectedPoint := NewArea(point.Add(offset)).Build()
			require.Equal(t, expectedPoint, newPoint, "%s+%s\nexpected=%s\nactual=%s", point.String(), offset.String(), expectedPoint.String(), newPoint.String())
		}
	}
}

func TestRotate(t *testing.T) {
	for i := 0; i < 5; i++ {
		for q := int64(-2); q < 2; q++ {
			for r := int64(-2); r < 2; r++ {
				orig := hex.Hex{Q: 1, R: 1}
				pivot := hex.Hex{Q: q, R: r}
				rotatedArea := NewArea(orig).Rotate(pivot, i).Build()
				var found hex.Hex
				for _, h := range rotatedArea.Slice() {
					found = h
					break
				}
				expected := orig.Rotate(pivot, i)
				dbg := NewArea(orig, pivot, found, expected).String()
				require.Equal(t, expected, found, "rotated %s about %s by %d. %s", orig.String(), pivot.String(), i, dbg)
			}
		}
	}
}

func TestRotateNOP(t *testing.T) {
	for i := 0; i < 5; i++ {
		for q := int64(-2); q < 2; q++ {
			for r := int64(-2); r < 2; r++ {
				orig := BigHex(hex.Origin(), 4).Build()
				pivot := hex.Hex{Q: q, R: r}
				if (pivot == hex.Hex{} || i == 0) {

					rotatedArea := orig.Rotate(pivot, i).Build()
					require.True(t, orig.Equals(rotatedArea), "NOP rotate")

				}
			}
		}
	}
}

func TestTriangle(t *testing.T) {
	// points for a big triangle
	points := []hex.Hex{
		{Q: 1, R: -2},
		{Q: 1, R: 1},
		{Q: -2, R: 1},
	}
	outline := Line(append(points, points[0])...)
	expectedOutline := []hex.Hex{
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
	assert.True(t, expectedOutlineArea.Equals(outline), "expected=%s\nactual=  %s\n", expectedOutlineArea.String(), outline.String())

	fill := Polygon(points...).Build()
	expectedFillArea := NewArea(append(expectedOutline, hex.Origin())...)
	assert.True(t, expectedFillArea.Equals(fill), "expected=%s\nactual=  %s\n", expectedFillArea.String(), fill.String())
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

		// rotate
		assert.Equal(t, b.expected, b.a.Rotate(hex.Hex{Q: 10, R: -10}, 3).Build().CheckBounding(b.b.Rotate(hex.Hex{Q: 10, R: -10}, 3).Build()), "\na=%s\nb=%s", b.a.String(), b.b.String())

		// translate
		assert.Equal(t, b.expected, b.a.Translate(hex.Hex{Q: -3, R: 100}).Build().CheckBounding(b.b.Translate(hex.Hex{Q: -3, R: 100}).Build()), "\na=%s\nb=%s", b.a.String(), b.b.String())
	})
}

func TestBounding(t *testing.T) {
	tests := []boundTest{
		{a: NewArea(hex.Origin()), b: NewArea(hex.Origin()), expected: Equals},
		{a: NewArea(), b: NewArea(hex.Origin()), expected: Undefined},
		{a: BigHex(hex.Origin(), 4), b: NewArea(hex.Origin()), expected: Contains},
		{a: BigHex(hex.Origin(), 4), b: NewArea(hex.Hex{Q: 100, R: 100}), expected: Distinct},
		{a: BigHex(hex.Origin(), 4), b: BigHex(hex.Hex{Q: 1, R: 1}, 4), expected: Overlap},
		{a: BigHex(hex.Origin(), 5), b: BigHex(hex.Origin(), 5).Subtract(NewArea(hex.Hex{Q: 1, R: 1})).Build(), expected: Contains},
		{a: Rectangle(hex.Hex{Q: 5, R: 5}, hex.Hex{Q: 10, R: 10}).Union(NewArea(hex.Origin())).Build(), b: NewArea(hex.Origin()), expected: Contains},
	}
	for i, test := range tests {
		test.assertBound(t, fmt.Sprintf("%d", i))
	}
}
