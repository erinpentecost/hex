package curve

import (
	"image"
	"image/color"
	"math"

	"github.com/erinpentecost/hexcoord/pos"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Decorator interface {
	AreaColor(h pos.Hex) color.RGBA
	EdgeColor(h pos.Hex, dir int) color.RGBA
	AreaLabel(h pos.Hex) string
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{200, 100, 0, 255}
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

// renderGrid draws a hex grid.
// imageLen is the width and height of the image.
// cameraDist is some number between 0 and 1. The closer to 0,
// one hex will fill the viewing pane. The closer to 1, the more
// hexes will fill the pahe. 1/zoom is the number of hexes along
// one axis that will be rendered.
// center is the hex coord that is at the center of the image.
// d is a Decorator that describes the colors and labels for each hex.
func renderGrid(imageLen int, zoom float64, center pos.Hex, d Decorator) *image.RGBA {
	if zoom > 1.0 || zoom <= 0.0 {
		panic("zoom range is from 0 to 1")
	}

	m := image.NewRGBA(image.Rect(0, 0, imageLen, imageLen))

	// determine how many hexes are in the image
	hexDimension := math.Ceil(1.0 / zoom)

	done := make(chan interface{})
	defer close(done)
	drawableHexes := center.RectangleArea(done, opposite Hex)

	for tile := range work {
		for x := tile.x1; x < tile.x2; x++ {
			for y := tile.y1; y < tile.y2; y++ {
				//setColor(m, colors, x, y, i, zoom)
			}
		}

	}

	return m

}
