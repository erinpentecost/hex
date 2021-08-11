package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/erinpentecost/hex"
	"github.com/erinpentecost/hex/area"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var outFile string
var width int

func init() {
	flag.StringVar(&outFile, "file", "", "png file to save the image to.")

	flag.IntVar(&width, "w", 500, "width of the image")

}

type annotatedHex struct {
	hex.Hex
	Label string
}

func main() {
	flag.Parse()
	var fileHandle *os.File
	var err error
	if outFile == "" {
		fileHandle, err = os.CreateTemp(os.TempDir(), "hex-*.png")
		if err != nil {
			log.Fatal(fmt.Sprintf("failed to make temp file: %v", err))
		}
	} else {
		outFile, err = filepath.Abs(outFile)
		if err != nil {
			log.Fatal(fmt.Sprintf("failed to find path: %v", err))
		}
		fileHandle, err = os.Create(outFile)
		if err != nil {
			log.Fatal(fmt.Sprintf("failed to make file %s: %v", outFile, err))
		}
	}

	os.Stderr.WriteString(fmt.Sprintf("outfile=%s, width=%d\n", fileHandle.Name(), width))

	var wrappedHexes []annotatedHex

	os.Stderr.WriteString("reading from stdin...\n")

	err = json.NewDecoder(os.Stdin).Decode(&wrappedHexes)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to decode stdin: %v", err))
	}

	os.Stderr.WriteString("decoded json. drawing...\n")

	labelLookup := make(map[hex.Hex]string)
	hexes := make([]hex.Hex, len(wrappedHexes))
	for i, h := range wrappedHexes {
		hexes[i] = h.Hex
		labelLookup[h.Hex] = h.Label
	}

	cc := NewCamera(width, area.NewArea(hexes...), func(h hex.Hex) string { return labelLookup[h] })
	img := cc.Draw()

	os.Stderr.WriteString("saving...\n")

	err = Save(img, fileHandle)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to save image to file: %v", err))
	}

	os.Stdout.WriteString(fileHandle.Name())
	os.Stdout.WriteString("\n")
}

// Camera defines the viewing pane
type Camera struct {
	imageLenX int
	imageLenY int
	centerX   float64
	centerY   float64
	hWidth    float64
	minR      int64
	maxR      int64
	minQ      int64
	maxQ      int64

	labeller func(hex.Hex) string
	area     *area.Area
}

// NewCamera creates a new camera object
func NewCamera(width int, area *area.Area, labeller func(hex.Hex) string) Camera {

	// find world bounds
	minR, maxR, minQ, maxQ, err := area.Bounds()
	if err != nil {
		log.Fatal(err)
	}
	//minR = minR - 2
	//maxR = maxR + 2

	// world bounds
	topLeftX, topLeftY := hex.Hex{Q: minQ, R: minR}.ToHexFractional().ToCartesian()
	bottomRightX, bottomRightY := hex.Hex{Q: maxQ, R: maxR}.ToHexFractional().ToCartesian()
	_, addY := hex.Hex{Q: -1, R: 2}.ToHexFractional().ToCartesian()
	addY = math.Abs(addY) * 1.5
	worldX := bottomRightX - topLeftX
	worldY := bottomRightY - topLeftY + addY

	worldDiag := math.Sqrt(float64(worldX*worldX + worldY*worldY))

	// height is determined by aspect ratio of world space
	height := int(float64(width) * worldY / worldX)

	imageLenX := width
	imageLenY := height

	imgDiag := math.Sqrt(float64(imageLenX*imageLenX + imageLenY*imageLenY))

	centerX, centerY := hex.LerpHex(hex.Hex{Q: minQ, R: minR}, hex.Hex{Q: maxQ, R: maxR}, 0.5).ToHexFractional().ToCartesian()

	hWidth := imgDiag / worldDiag

	wrappedLabeller := func(h hex.Hex) string {
		if l := labeller(h); l != "" {
			return l
		}
		return h.String()
	}

	return Camera{
		imageLenX: imageLenX,
		imageLenY: imageLenY,
		centerX:   centerX,
		centerY:   centerY,
		hWidth:    hWidth,
		minR:      minR,
		maxR:      maxR,
		minQ:      minQ,
		maxQ:      maxQ,
		area:      area,
		labeller:  wrappedLabeller,
	}
}

// ScreenToHex converts camera coordinates to hex coordinates
func (c Camera) ScreenToHex(x, y int) hex.HexFractional {
	xM := (float64(x-c.imageLenX/2) / c.hWidth) + c.centerX
	xY := (float64(y-c.imageLenY/2) / c.hWidth) + c.centerY
	return hex.HexFractionalFromCartesian(xM, xY)
}

// HexToScreen converts hex coord to screen coord.
// returned value may be out of bounds.
func (c Camera) HexToScreen(p hex.HexFractional) (x, y int) {
	hx, hy := p.ToCartesian()
	return int((hx-c.centerX)*c.hWidth) + c.imageLenX/2, int((hy-c.centerY)*c.hWidth) + c.imageLenY/2
}

// Grid draws a hex grid.
func (c Camera) Draw() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, c.imageLenX, c.imageLenY))
	seen := make(map[hex.Hex]interface{})
	// look at each pixel and color in the hex background
	for x := 0; x < c.imageLenX; x++ {
		for y := 0; y < c.imageLenY; y++ {
			hf := c.ScreenToHex(x, y)
			hd := hf.ToHex()
			img.SetRGBA(x, y, AreaColor(hd, c.area))
			seen[hd] = nil
		}
	}

	// label the hexes
	if c.hWidth > 75.0 {
		for h := range seen {
			col := color.RGBA{0, 0, 0, 255}
			if c.area.ContainsHexes(h) {
				col = color.RGBA{200, 200, 200, 255}
			}
			hx, hy := c.HexToScreen(h.ToHexFractional())
			addLabel(img, hx-int(c.hWidth)/2, hy, col, c.labeller(h))
		}
	}

	return img
}

func addLabel(img *image.RGBA, x, y int, col color.RGBA, label string) {
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
	img.SetRGBA(x, y, col)
}

// AreaColor picks a background color for the hex.
func AreaColor(h hex.Hex, note *area.Area) color.RGBA {
	if note.ContainsHexes(h) {
		return color.RGBA{50, 50, 50, 255}
	}

	mod := func(a int64) int64 {
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

// Save saves an image to a file
func Save(img *image.RGBA, f *os.File) error {
	defer f.Close()
	return png.Encode(f, img)
}
