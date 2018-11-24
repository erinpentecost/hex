package draw

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
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

// EdgeColor picks an edge color.
func (d DefaultDecorator) EdgeColor(h pos.Hex, dir int) color.RGBA {
	return color.RGBA{255, 255, 255, 255}
}

// AreaLabel uses the hex's coordinates.
func (d DefaultDecorator) AreaLabel(h pos.Hex) string {
	return h.ToString()
}

// Camera defines the viewing pane
type Camera struct {
	imageLenX int
	imageLenY int
	zoom      float64
	center    pos.Hex
	centerX   float64
	centerY   float64
	hWidth    float64
}

// NewCamera creates a new camera object
func NewCamera(imageLenX int, imageLenY int, zoom float64, center pos.Hex) Camera {
	if zoom > 1.0 || zoom <= 0.0 {
		panic("zoom range is from 0 to 1")
	}
	centerX, centerY := center.ToHexFractional().ToCartesian()
	diag := math.Sqrt(float64(imageLenX*imageLenX + imageLenY*imageLenY))
	return Camera{
		imageLenX: imageLenX,
		imageLenY: imageLenY,
		zoom:      zoom,
		center:    center,
		centerX:   centerX,
		centerY:   centerY,
		hWidth:    diag * zoom,
	}
}

// Scale returns the relation between screen coordinates and hex coordinates
func (c Camera) Scale() float64 {
	return c.hWidth
}

// ScreenToHex converts camera coordinates to hex coordinates
func (c Camera) ScreenToHex(x, y int) pos.HexFractional {
	xM := (float64(x-c.imageLenX/2) / c.hWidth) + c.centerX
	xY := (float64(y-c.imageLenY/2) / c.hWidth) + c.centerY
	return pos.HexFractionalFromCartesian(xM, xY)
}

// HexToScreen converts hex coord to screen coord.
// returned value may be out of bounds.
func (c Camera) HexToScreen(p pos.HexFractional) (x, y int) {
	hx, hy := p.ToCartesian()
	return int((hx-c.centerX)*c.hWidth) + c.imageLenX/2, int((hy-c.centerY)*c.hWidth) + c.imageLenY/2
}

// Render draws a hex grid.
func (c Camera) Render(d Decorator) *image.RGBA {

	m := image.NewRGBA(image.Rect(0, 0, c.imageLenX, c.imageLenY))

	foundHexes := make(map[pos.Hex]interface{})

	// look at each pixel and color in the hex background
	for x := 0; x < c.imageLenX; x++ {
		for y := 0; y < c.imageLenY; y++ {
			hf := c.ScreenToHex(x, y)
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
		if label == "" {
			continue
		}
		hx, hy := c.HexToScreen(h.ToHexFractional())
		areaCol := d.AreaColor(h)
		invertCol := color.RGBA{areaCol.R ^ 0xFF, areaCol.G ^ 0xFF, areaCol.B ^ 0xFF, 255}
		addLabel(m, hx, hy, invertCol, label)
	}

	// todo: draw edges

	return m
}

func addLabel(img *image.RGBA, x, y int, col color.RGBA, label string) {
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

// Point draws a fat point.
func (c Camera) Point(img *image.RGBA, col color.RGBA, p pos.HexFractional) {
	xImg, yImg := c.HexToScreen(p)

	img.SetRGBA(xImg, yImg, col)

	img.SetRGBA(xImg+1, yImg, col)
	img.SetRGBA(xImg-1, yImg, col)
	img.SetRGBA(xImg, yImg+1, col)
	img.SetRGBA(xImg, yImg-1, col)

	img.SetRGBA(xImg+1, yImg+1, col)
	img.SetRGBA(xImg-1, yImg-1, col)
	img.SetRGBA(xImg-1, yImg+1, col)
	img.SetRGBA(xImg+1, yImg-1, col)
}

// Line draws a line on the image.
func (c Camera) Line(img *image.RGBA, col color.RGBA, bold bool, start, end pos.HexFractional) {
	len := start.DistanceTo(end)
	sampleStep := float64(0.99) / (len * c.Scale())
	for s := 0.0; s < 1.0; s = s + sampleStep {
		posHex := pos.LerpHexFractional(start, end, s)
		if bold {
			c.Point(img, col, posHex)
		} else {
			xImg, yImg := c.HexToScreen(posHex)
			img.SetRGBA(xImg, yImg, col)
		}
	}
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
