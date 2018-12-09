package draw

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/erinpentecost/hexcoord/curve"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Decorator informs the renderer how to draw the hex.
type Decorator interface {
	AreaColor(h pos.Hex) color.RGBA
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

// AreaLabel uses the hex's coordinates.
func (d DefaultDecorator) AreaLabel(h pos.Hex) string {
	return h.ToString()
}

// UnlabeledDecorator does minimal styling.
type UnlabeledDecorator struct{}

// AreaColor picks a background color for the hex.
func (d UnlabeledDecorator) AreaColor(h pos.Hex) color.RGBA {
	mod := func(a int) int {
		if a < 0 {
			return (a * (-1)) % 2
		}
		return a % 2
	}

	m := mod(h.Q) + 2*mod(h.R)
	switch m {
	case 0:
		return color.RGBA{255, 222, 222, 255}
	case 1:
		return color.RGBA{222, 255, 222, 255}
	case 2:
		return color.RGBA{222, 222, 255, 255}
	default:
		return color.RGBA{255, 255, 222, 255}
	}
}

// AreaLabel uses the hex's coordinates.
func (d UnlabeledDecorator) AreaLabel(h pos.Hex) string {
	return ""
}

// Camera defines the viewing pane
type Camera struct {
	img       *image.RGBA
	imageLenX int
	imageLenY int
	zoom      float64
	center    pos.Hex
	centerX   float64
	centerY   float64
	hWidth    float64
}

// NewCamera creates a new camera object
func NewCamera(img *image.RGBA, zoom float64, center pos.Hex) Camera {
	if zoom > 1.0 || zoom <= 0.0 {
		panic("zoom range is from 0 to 1")
	}
	centerX, centerY := center.ToHexFractional().ToCartesian()
	imageLenX := img.Rect.Dx()
	imageLenY := img.Rect.Dy()
	diag := math.Sqrt(float64(imageLenX*imageLenX + imageLenY*imageLenY))
	return Camera{
		img:       img,
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

// Grid draws a hex grid.
func (c Camera) Grid(d Decorator) {

	foundHexes := make(map[pos.Hex]interface{})

	// look at each pixel and color in the hex background
	for x := 0; x < c.imageLenX; x++ {
		for y := 0; y < c.imageLenY; y++ {
			hf := c.ScreenToHex(x, y)
			hd := hf.ToHex()
			if _, ok := foundHexes[hd]; !ok {
				foundHexes[hd] = nil
			}
			c.img.SetRGBA(x, y, d.AreaColor(hd))
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
		addLabel(c.img, hx, hy, invertCol, label)
	}

	// todo: draw edges
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
func (c Camera) Point(to color.RGBA, p pos.HexFractional) {
	xImg, yImg := c.HexToScreen(p)

	col, _ := colorful.MakeColor(to)

	// center point, full strength
	c.img.SetRGBA(xImg, yImg, to)

	orth := 0.6
	blend(c.img, xImg+1, yImg, col, orth)
	blend(c.img, xImg-1, yImg, col, orth)
	blend(c.img, xImg, yImg+1, col, orth)
	blend(c.img, xImg, yImg-1, col, orth)

	diag := 0.1
	blend(c.img, xImg+1, yImg+1, col, diag)
	blend(c.img, xImg-1, yImg-1, col, diag)
	blend(c.img, xImg-1, yImg+1, col, diag)
	blend(c.img, xImg+1, yImg-1, col, diag)
}

func blend(img *image.RGBA, x, y int, col colorful.Color, strength float64) {
	from, ok := colorful.MakeColor(img.RGBAAt(x, y))
	if ok {
		end := from.BlendLab(col, strength).Clamped()
		img.Set(x, y, end)
	} else {
		img.Set(x, y, col.Clamped())
	}
}

// Curve draws curve on the image.
func (c Camera) Curve(col color.RGBA, curver curve.Curver) {
	// Draw tangent lines
	supportLen := 0.5
	initPoint, initTan, _ := curver.Sample(0.0)
	endPoint, endTan, _ := curver.Sample(1.0)
	//midPoint, midTan, _ := curver.Sample(0.5)
	c.Line(color.RGBA{255, 0, 0, 255}, false, initPoint, initPoint.Add(initTan.Normalize().Multiply(supportLen)))
	c.Line(color.RGBA{0, 0, 255, 255}, false, endPoint, endPoint.Add(endTan.Normalize().Multiply(supportLen)))
	//c.Line(color.RGBA{0, 255, 255, 255}, false, midPoint, midPoint.Add(midTan.Normalize().Multiply(supportLen)))

	//c.Line(color.RGBA{0, 255, 255, 255}, false, initPoint, initCurva.Add(initPoint))
	//c.Line(color.RGBA{0, 255, 255, 255}, false, endPoint, endCurva.Add(endPoint))

	// Trace curve
	sampleStep := float64(0.99) / (curver.Length() * c.Scale())
	for s := 0.0; s < 1.0; s = s + sampleStep {
		posHex, _, _ := curver.Sample(s)
		c.Point(col, posHex)
	}
}

// Line draws a line on the image.
func (c Camera) Line(col color.RGBA, bold bool, start, end pos.HexFractional) {
	len := start.DistanceTo(end)
	sampleStep := float64(0.99) / (len * c.Scale())
	for s := 0.0; s < 1.0; s = s + sampleStep {
		posHex := pos.LerpHexFractional(start, end, s)
		if bold {
			c.Point(col, posHex)
		} else {
			xImg, yImg := c.HexToScreen(posHex)
			c.img.SetRGBA(xImg, yImg, col)
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
