package example_test

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/erinpentecost/hexcoord/curve"
	"github.com/erinpentecost/hexcoord/draw"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
)

func getColor(c curve.Curver) color.RGBA {
	switch c.Spin() {
	case curve.Clockwise:
		return color.RGBA{0, 0, 222, 255}
	case curve.CounterClockwise:
		return color.RGBA{222, 0, 0, 255}
	default:
		return color.RGBA{0, 0, 0, 255}
	}
}

func TestSmoothCurveDrawing(t *testing.T) {
	ti := pos.HexFractional{Q: 1, R: -1}.Normalize()
	te := pos.HexFractional{Q: -1, R: 1}.Normalize()
	path := []pos.HexFractional{
		pos.OriginFractional(),
		pos.HexFractional{Q: 1, R: -1},
		pos.HexFractional{Q: 1, R: 0},
		pos.HexFractional{Q: 0, R: 1},
		pos.HexFractional{Q: 1, R: 1},
		pos.HexFractional{Q: 2, R: 0},
		pos.HexFractional{Q: 2, R: -1},
	}

	smoothArcs := curve.SmoothPath(ti, te, path)

	dd := draw.DefaultDecorator{}
	img := image.NewRGBA(image.Rect(0, 0, 600, 600))
	cc := draw.NewCamera(img, 0.06, pos.Hex{Q: 1, R: 0})

	cc.Grid(dd)

	// Draw arcs.
	for _, arc := range smoothArcs {
		curve := arc.Curve()
		cc.Curve(getColor(curve), curve)
	}

	fpath, err := draw.Save(img, "TestSmoothCurveDrawing.png")
	assert.NoError(t, err, fpath)
	fmt.Println(fpath)

}

func TestBiarcDrawing(t *testing.T) {
	dd := draw.DefaultDecorator{}
	img := image.NewRGBA(image.Rect(0, 0, 500, 400))
	cc := draw.NewCamera(img, 0.15, pos.Origin())

	cc.Grid(dd)

	rVals := []float64{0.5, 1.0, 1.5}
	for _, r := range rVals {
		b := curve.Biarc(
			pos.HexFractional{Q: -1.0, R: 0.0},
			pos.HexFractional{Q: 1.0, R: -2.1},
			pos.HexFractional{Q: 1.0, R: 0.0},
			pos.HexFractional{Q: 1.0, R: -0.1},
			r)
		for _, arc := range b {
			c := arc.Curve()
			cc.Curve(getColor(c), c)
		}
	}

	fpath, err := draw.Save(img, "TestBiarcDrawing.png")
	assert.NoError(t, err, fpath)
	fmt.Println(fpath)
}

func TestHappyFaceDrawing(t *testing.T) {

	dd := draw.DefaultDecorator{}
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	cc := draw.NewCamera(img, 0.1, pos.Hex{Q: 1, R: 0})

	cc.Grid(dd)

	// mouth
	clockwiseArc := curve.CircularArc{
		I: pos.HexFractional{Q: 2, R: 0},
		T: pos.HexFractional{Q: -1, R: 2},
		E: pos.HexFractional{Q: 0, R: 0},
	}.Curve()

	// left eye
	counterclockwiseArc := curve.CircularArc{
		I: pos.HexFractional{Q: 1, R: -1},
		T: pos.HexFractional{Q: 1, R: -2},
		E: pos.HexFractional{Q: 0, R: -1},
	}.Curve()

	// right eye, wink
	lineArc := curve.CircularArc{
		I: pos.HexFractional{Q: 2, R: -1},
		T: pos.HexFractional{Q: 1, R: 0},
		E: pos.HexFractional{Q: 3, R: -1},
	}.Curve()

	cc.Curve(getColor(clockwiseArc), clockwiseArc)
	cc.Curve(getColor(counterclockwiseArc), counterclockwiseArc)
	cc.Curve(getColor(lineArc), lineArc)

	fpath, err := draw.Save(img, "TestHappyFaceDrawing.png")
	assert.NoError(t, err, fpath)
	fmt.Println(fpath)
}
