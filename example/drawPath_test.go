package example_test

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/erinpentecost/hexcoord/curve"
	"github.com/erinpentecost/hexcoord/draw"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
)

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
	cc := draw.NewCamera(600, 500, 0.15, pos.Hex{Q: 1, R: 0})

	img := cc.Render(dd)

	getColor := func(i int, spin bool) (line, support color.RGBA) {
		r := uint8((i%2)*10) + 50
		b := uint8(255)
		if spin {
			b = 100
		}
		return color.RGBA{r, 0, b, 255}, color.RGBA{r, 100, b, 255}
	}

	// Draw supporting vectors.
	for i, arc := range smoothArcs {
		curve := arc.Curve()
		s := curve.Spin() == 1
		_, supportColor := getColor(i, s)

		supportLen := 0.5
		initPoint, initTan, _ := curve.Sample(0.0)
		endPoint, endTan, _ := curve.Sample(1.0)
		cc.Line(img, supportColor, false, initPoint, initPoint.Add(initTan.Normalize().Multiply(supportLen)))
		cc.Line(img, supportColor, false, endPoint, endPoint.Add(endTan.Normalize().Multiply(supportLen)))
	}

	// Draw arcs.
	for i, arc := range smoothArcs {
		curve := arc.Curve()
		s := curve.Spin() == 1
		sampleStep := float64(0.99) / (curve.Length() * cc.Scale())
		arcColor, _ := getColor(i, s)
		for s := 0.0; s < 1.0; s = s + sampleStep {
			posHex, _, _ := curve.Sample(s)
			cc.Point(img, arcColor, posHex)
		}
	}

	fpath, err := draw.Save(img, "TestSmoothCurveDrawing.png")
	assert.NoError(t, err, fpath)
	fmt.Println(fpath)

}

func TestHappyFaceDrawing(t *testing.T) {

	dd := draw.DefaultDecorator{}
	cc := draw.NewCamera(600, 500, 0.1, pos.Hex{Q: 1, R: 0})

	img := cc.Render(dd)

	// mouth
	clockwiseArc := curve.CircularArc{
		I: pos.HexFractional{Q: 2, R: 0},
		T: pos.HexFractional{Q: -1, R: 2},
		E: pos.HexFractional{Q: 0, R: 0},
	}

	// left eye
	counterclockwiseArc := curve.CircularArc{
		I: pos.HexFractional{Q: 1, R: -1},
		T: pos.HexFractional{Q: 1, R: -2},
		E: pos.HexFractional{Q: 0, R: -1},
	}

	// right eye, wink
	lineArc := curve.CircularArc{
		I: pos.HexFractional{Q: 2, R: -1},
		T: pos.HexFractional{Q: 1, R: 0},
		E: pos.HexFractional{Q: 3, R: -1},
	}

	black := color.RGBA{0, 0, 0, 255}

	cc.Curve(img, black, clockwiseArc.Curve())
	cc.Curve(img, black, counterclockwiseArc.Curve())
	cc.Curve(img, black, lineArc.Curve())

	fpath, err := draw.Save(img, "TestHappyFaceDrawing.png")
	assert.NoError(t, err, fpath)
	fmt.Println(fpath)
}
