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
	img := image.NewRGBA(image.Rect(0, 0, 900, 900))
	cc := draw.NewCamera(img, 0.15, pos.Hex{Q: 1, R: 0})

	cc.Grid(dd)

	// Draw arcs.
	for _, arc := range smoothArcs {
		curve := arc.Curve()
		col := getColor(curve)
		cc.Curve(col, curve)
		if !pos.OriginFractional().AlmostEquals(arc.C) {
			colc := color.RGBA{100, 100, 100, 255}
			cc.Line(colc, false, arc.I, arc.C)
			cc.Line(colc, false, arc.E, arc.C)
		}
	}

	fpath, err := draw.Save(img, "TestSmoothCurveDrawing.png")
	assert.NoError(t, err, fpath)
	fmt.Println(fpath)

}

func TestBiarcDrawing(t *testing.T) {
	center := pos.HexFractional{Q: -100.0, R: 100.0}
	dd := draw.DefaultDecorator{}
	img := image.NewRGBA(image.Rect(0, 0, 500, 400))
	cc := draw.NewCamera(img, 0.15, center.ToHex())

	cc.Grid(dd)

	left := pos.HexFractional{Q: -1.0, R: 0.0}.Add(center)
	right := pos.HexFractional{Q: 1.0, R: 0.0}.Add(center)
	upish := pos.HexFractional{Q: 1.0, R: -2.0}.Normalize()
	rightish := pos.HexFractional{Q: 1.0, R: 0.0}.Normalize()

	rVals := []float64{10.1}
	for _, r := range rVals {
		top := curve.Biarc(
			left,
			upish,
			right,
			rightish,
			r)
		for _, arc := range top {
			c := arc.Curve()
			col := getColor(c)
			cc.Curve(col, c)
		}

		/*bottom := curve.Biarc(
			right,
			rightish.Rotate(pos.OriginFractional(), math.Pi),
			left,
			upish.Rotate(pos.OriginFractional(), math.Pi),
			r)
		for _, arc := range bottom {
			c := arc.Curve()
			col := getColor(c)
			cc.Curve(col, c)
		}*/
	}

	fpath, err := draw.Save(img, "TestBiarcDrawing.png")
	assert.NoError(t, err, fpath)
	fmt.Println(fpath)
}

func TestBiarcHugeDrawing(t *testing.T) {
	center := pos.HexFractional{Q: -100.0, R: 100.0}
	scale := 10.0
	dd := draw.UnlabeledDecorator{}
	img := image.NewRGBA(image.Rect(0, 0, 500, 400))
	cc := draw.NewCamera(img, 1.0/(scale*scale), center.ToHex())

	cc.Grid(dd)

	left := pos.HexFractional{Q: -1.0, R: 0.0}.Multiply(scale).Add(center)
	right := pos.HexFractional{Q: 1.0, R: 0.0}.Multiply(scale).Add(center)
	upish := pos.HexFractional{Q: 1.0, R: -2.0}.Normalize()
	rightish := pos.HexFractional{Q: 1.0, R: 0.0}.Normalize()

	rVals := []float64{10.1}
	for _, r := range rVals {
		top := curve.Biarc(
			left,
			upish,
			right,
			rightish,
			r)
		for _, arc := range top {
			c := arc.Curve()
			col := getColor(c)
			cc.Curve(col, c)
		}
	}

	fpath, err := draw.Save(img, "TestBiarcHugeDrawing.png")
	assert.NoError(t, err, fpath)
	fmt.Println(fpath)
}

func TestHappyFaceDrawingWithBiarcs(t *testing.T) {

	dd := draw.DefaultDecorator{}
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	cc := draw.NewCamera(img, 0.1, pos.Hex{Q: 1, R: 0})

	cc.Grid(dd)

	drawBiarc := func(arcs []curve.CircularArc) {
		for _, arc := range arcs {
			c := arc.Curve()
			cc.Curve(getColor(c), c)
		}
	}

	// mouth
	clockwiseArc := curve.Biarc(
		pos.HexFractional{Q: 2, R: 0},
		pos.HexFractional{Q: -1, R: 2},
		pos.HexFractional{Q: 0, R: 0},
		pos.HexFractional{Q: 1, R: -2},
		1.0)
	drawBiarc(clockwiseArc)

	// left eye
	counterclockwiseArc := curve.Biarc(
		pos.HexFractional{Q: 1, R: -1},
		pos.HexFractional{Q: 1, R: -2},
		pos.HexFractional{Q: 0, R: -1},
		pos.HexFractional{Q: -1, R: 2},
		1.0)
	drawBiarc(counterclockwiseArc)

	// right eye, wink
	lineArc := curve.Biarc(
		pos.HexFractional{Q: 2, R: -1},
		pos.HexFractional{Q: 1, R: 0},
		pos.HexFractional{Q: 3, R: -1},
		pos.HexFractional{Q: 1, R: 0},
		1.0)
	drawBiarc(lineArc)

	fpath, err := draw.Save(img, "TestHappyFaceDrawingWithBiarcs.png")
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
