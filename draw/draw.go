package draw

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/erinpentecost/hexcoord/pos"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Decorator informs the renderer how to draw the hex.
type Decorator interface {
	AreaColor(h pos.Hex) color.RGBA
	EdgeColor(h pos.Hex, dir int) color.RGBA
	AreaLabel(h pos.Hex) string
}

// DefaultDecorator draws a boring hex map.
type DefaultDecorator struct{}

// AreaColor picks a background color for the hex.
func (d DefaultDecorator) AreaColor(h pos.Hex) color.RGBA {
	m := (h.Q % 2) + 2*(h.R%2)
	switch m {
	case 0:
		return color.RGBA{100, 0, 0, 255}
	case 1:
		return color.RGBA{0, 100, 0, 255}
	case 2:
		return color.RGBA{0, 0, 100, 255}
	default:
		return color.RGBA{0, 100, 100, 255}
	}
}

// EdgeColor picks an edge color.
func (d DefaultDecorator) EdgeColor(h pos.Hex, dir int) color.RGBA {
	return color.RGBA{255, 255, 255, 255}
}

// AreaLabel uses the hex's coordinates.
func (d DefaultDecorator) AreaLabel(h pos.Hex) string {
	return h.ToString()
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
	img.SetRGBA(x, y, col)
}

// RenderGrid draws a hex grid.
// imageLen is the width and height of the image.
// zoom is some number between 0 and 1. The closer to 1,
// one hex will fill the viewing pane. The closer to 0, the more
// hexes will fill the pane.
// center is the hex coord that is at the center of the image.
// d is a Decorator that describes the colors and labels for each hex.
func RenderGrid(imageLen int, zoom float64, center pos.Hex, d Decorator) *image.RGBA {
	if zoom > 1.0 || zoom <= 0.0 {
		panic("zoom range is from 0 to 1")
	}

	m := image.NewRGBA(image.Rect(0, 0, imageLen, imageLen))
	centerX, centerY := center.ToHexFractional().ToCartesian()

	screenToHex := func(x, y int) pos.HexFractional {
		hWidth := float64(imageLen) * zoom
		xM := (float64(x) / hWidth) + centerX
		xY := (float64(y) / hWidth) + centerY
		return pos.HexFractionalFromCartesian(xM, xY)
	}
	// hexToScreen converts hex coord to screen coord.
	// returned value may be out of bounds.
	hexToScreen := func(p pos.HexFractional) (x, y int) {
		hWidth := float64(imageLen) * zoom
		hx, hy := p.ToCartesian()
		return int((hx - centerX) * hWidth), int((hy - centerY) * hWidth)
	}

	foundHexes := make(map[pos.Hex]interface{})

	// look at each pixel and color in the hex background
	for x := 0; x < imageLen; x++ {
		for y := 0; y < imageLen; y++ {
			hf := screenToHex(x, y)
			hd := hf.ToHex()
			if _, ok := foundHexes[hd]; !ok {
				foundHexes[hd] = nil
			}
			m.SetRGBA(x, y, d.AreaColor(hd))
		}
	}

	// label the hexes
	for h := range foundHexes {
		label := d.AreaLabel(h)
		hy, hx := hexToScreen(h.ToHexFractional())
		addLabel(m, hx, hy, label)
	}

	// draw the edges
	// todo

	return m
}

// Save saves an image to a file
func Save(img *image.RGBA, path string) (string, error) {
	fullPath := fmt.Sprintf("%s", path)
	f, err := os.Create(path)
	if err != nil {
		return fullPath, err
	}
	defer f.Close()
	png.Encode(f, img)
	return fullPath, nil
}
