package draw_test

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/erinpentecost/hexcoord/draw"

	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
)

func TestDraw(t *testing.T) {
	dd := draw.DefaultDecorator{}
	img := image.NewRGBA(image.Rect(0, 0, 500, 600))
	cc := draw.NewCamera(img, 0.2, pos.Origin())

	cc.Grid(dd)

	path, err := draw.Save(img, "testdraw.png")
	assert.NoError(t, err, path)
	fmt.Println(path)
}

type HighlightDecorator struct {
	interest map[pos.Hex]interface{}
}

// AreaColor picks a background color for the hex.
func (d HighlightDecorator) AreaColor(h pos.Hex) color.RGBA {
	_, interesting := d.interest[h]
	if interesting {
		return color.RGBA{0, 0, 0, 255}
	}

	mod := func(a int) int {
		if a < 0 {
			return (a * (-1)) % 2
		}
		return a % 2
	}

	m := mod(h.Q) + 2*mod(h.R)
	switch m {
	case 0:
		return color.RGBA{255, 200, 200, 255}
	case 1:
		return color.RGBA{200, 255, 200, 255}
	case 2:
		return color.RGBA{200, 200, 255, 255}
	default:
		return color.RGBA{255, 255, 200, 255}
	}
}

// AreaLabel uses the hex's coordinates.
func (d HighlightDecorator) AreaLabel(h pos.Hex) string {
	return ""
}

func createLogoPoints() map[pos.Hex]interface{} {
	h := []pos.Hex{
		pos.Hex{Q: 2, R: -2},
		pos.Hex{Q: 1, R: -1},
		pos.Hex{Q: 0, R: 0},
		pos.Hex{Q: -1, R: 1},
		pos.Hex{Q: 1, R: 0},
		pos.Hex{Q: 1, R: 1},
	}
	ec := []pos.Hex{
		pos.Hex{Q: 2, R: -1},
		pos.Hex{Q: 1, R: -1},
		pos.Hex{Q: 0, R: 0},
		pos.Hex{Q: 0, R: 1},
		pos.Hex{Q: 1, R: 1},
	}
	o := []pos.Hex{
		pos.Hex{Q: 2, R: 0},
		pos.Hex{Q: 2, R: -1},
		pos.Hex{Q: 1, R: -1},
		pos.Hex{Q: 0, R: 0},
		pos.Hex{Q: 0, R: 1},
		pos.Hex{Q: 1, R: 1},
	}
	x := []pos.Hex{
		pos.Hex{Q: 2, R: -1},
		pos.Hex{Q: 1, R: 0},
		pos.Hex{Q: 0, R: 1},
		pos.Hex{Q: 1, R: -1},
		pos.Hex{Q: 1, R: 1},
	}
	r := []pos.Hex{
		pos.Hex{Q: 2, R: 0},
		pos.Hex{Q: 2, R: -1},
		pos.Hex{Q: 1, R: -1},
		pos.Hex{Q: 0, R: 0},
		pos.Hex{Q: -1, R: 1},
	}
	d := []pos.Hex{
		pos.Hex{Q: 2, R: 0},
		pos.Hex{Q: 2, R: -1},
		pos.Hex{Q: 1, R: -1},
		pos.Hex{Q: 0, R: 0},
		pos.Hex{Q: 0, R: 1},
		pos.Hex{Q: 1, R: 1},
		pos.Hex{Q: 3, R: -1},
		pos.Hex{Q: 4, R: -2},
	}

	logo := [][]pos.Hex{
		h,
		ec,
		x,
		ec,
		o,
		o,
		r,
		d,
	}

	taggedPos := make(map[pos.Hex]interface{})

	for offset, char := range logo {
		done := make(chan interface{})
		defer close(done)
		charOffset := pos.AreaMap(done, pos.Area(char...), func(x pos.Hex) pos.Hex { return x.Add(pos.Hex{Q: offset * 4, R: 0}) })
		for charSpot := range charOffset {
			taggedPos[charSpot] = nil
		}
	}

	return taggedPos
}

func TestDrawLogo(t *testing.T) {

	points := createLogoPoints()

	dd := HighlightDecorator{interest: points}
	img := image.NewRGBA(image.Rect(0, 0, 500, 100))
	cc := draw.NewCamera(img, 0.013, pos.Hex{Q: 15, R: 0})

	cc.Grid(dd)

	path, err := draw.Save(img, "TestDrawLogo.png")
	assert.NoError(t, err, path)
	fmt.Println(path)
}
