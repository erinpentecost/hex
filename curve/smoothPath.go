package curve

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/erinpentecost/hexcoord/pos"
)

// CircularArc defines a circular arc in three vectors.
type CircularArc struct {
	// I is the initial point.
	I pos.HexFractional
	// T is the tangent unit vector at the initial point.
	T pos.HexFractional
	// E is the end point.
	E pos.HexFractional
}

// ToString converts the arc to a string.
func (ca CircularArc) ToString() string {
	return fmt.Sprintf("{I: %v, T: %v, E: %v}", ca.I.ToString(), ca.T.ToString(), ca.E.ToString())
}

// SmoothPath takes as input a slice of connected Hexes.
// As output, it will return a piecewise collection of circular
// arcs that connect those hexes with G1 continuity.
// These arcs can be converted to parameterized curves with
// the Curve() function.
func SmoothPath(ti pos.HexFractional, te pos.HexFractional, path []pos.HexFractional) []CircularArc {

	// http://kaj.uniwersytetradom.pl/prace/Biarcs.pdf
	// https://en.wikipedia.org/wiki/Arc_length
	// https://en.wikipedia.org/wiki/Ramer%E2%80%93Douglas%E2%80%93Peucker_algorithm
	// https://www.redblobgames.com/articles/curved-paths/
	// http://www.ryanjuckett.com/programming/biarc-interpolation/
	// https://stag-ws.zcu.cz/ws/services/rest/kvalifikacniprace/downloadPraceContent?adipIdno=17817
	// https://www.ajdesigner.com/phpcircle/circle_arc_length_s.php

	// If there are 1 or fewer points, we are already
	// at the target path.
	if len(path) < 2 {
		return make([]CircularArc, 0, 0)
	}

	curves := make([]CircularArc, 0, 2*len(path))

	// Find tangents for each position.
	tangents := make([]pos.HexFractional, len(path), len(path))
	tangents[0] = ti
	tangents[len(path)-1] = te
	for p := 1; p < len(tangents)-1; p++ {
		tangents[p] = approximateTangent(path[p-1], path[p], path[p+1])
	}

	// Generate biarcs for each pair of points.
	for i := 0; i < len(path)-1; i++ {
		for _, b := range Biarc(path[i], tangents[i], path[i+1], tangents[i+1], 1.0) {
			curves = append(curves, b)
		}
	}

	return curves
}

// This algorithm was adapted from "The use of Piecewise Circular Curves in Geometric
// Modeling" by Ulugbek Khudayarov.
func approximateTangent(p0, p1, p2 pos.HexFractional) pos.HexFractional {
	a := p1.Subtract(p0)
	b := p2.Subtract(p1)
	aLen := a.Length()
	bLen := b.Length()

	return a.Multiply(bLen / aLen).Add(b.Multiply(aLen / bLen))
}

func findRoots(a, b, c complex128) (r1 complex128, r2 complex128) {
	component := cmplx.Sqrt(cmplx.Pow(b, 2) - 4.0*a*c)
	r1 = (-b + component) / (2.0 * a)
	r2 = (-b - component) / (2.0 * a)
	return
}

func chooseRoot(r1, r2 complex128) float64 {
	r1IsReal := closeEnough(imag(r1), 0.0)
	r2IsReal := closeEnough(imag(r2), 0.0)
	r1IsPositive := real(r1) >= 0.0
	r2IsPositive := real(r2) >= 0.0

	if r1IsReal && r2IsReal {
		if r1IsPositive && r2IsPositive {
			return math.Min(real(r1), real(r2))
		}
		return math.Max(real(r1), real(r2))
	}
	return math.Max(real(r1), real(r2))
}

func cartesianDotProduct(a, b pos.HexFractional) float64 {
	aX, aY := a.ToCartesian()
	bX, bY := b.ToCartesian()
	return aX*bX + aY*bY
}

// Biarc returns a list of circular arcs that connect pi to pe,
// with ti being the tangent at pi and te being the tangent at pe.
// This algorithm was adapted from "The use of Piecewise Circular Curves in Geometric
// Modeling" by Ulugbek Khudayarov.
func Biarc(pi, ti, pe, te pos.HexFractional, r float64) (arcs []CircularArc) {
	if r <= 0.0 {
		panic("r must be positive")
	}
	// Tangents should be unit vectors.
	ti = ti.Normalize()
	te = te.Normalize()

	//t := ti.Add(te)

	v := pi.Subtract(pe)

	// Single arc case
	_, pte, _ := CircularArc{pi, ti, pe}.Curve().Sample(1.0)
	if pte.AlmostEquals(te) {
		return []CircularArc{CircularArc{pi, ti, pe}}
	}

	// Now find the positive root for
	// v ⋅ v + 2 β v ⋅ ( r t s + t e ) + 2 r β 2 ( t s ⋅ t e − 1 ) = 0
	// β^2
	a := 2.0 * r * (cartesianDotProduct(ti, te) - 1.0)
	// β
	b := cartesianDotProduct(v.Multiply(2.0), ti.Multiply(r).Add(te))
	// constant
	c := cartesianDotProduct(v, v)

	// Semicircle case
	if closeEnough(a, 0.0) {
		// todo: needs to be fixed
		fmt.Println("semicircle case")
		j := pos.LerpHexFractional(pi, pe, 0.5)
		tj := ti.Multiply(-1.0)
		return []CircularArc{
			CircularArc{pi, ti, j},
			CircularArc{j, tj, pe},
		}
	}

	if closeEnough(cartesianDotProduct(ti, te), 1.0) {
		panic("unhandled special case #1")
	}
	if closeEnough(cartesianDotProduct(v, ti.Multiply(r).Add(te)), 0.0) {
		panic("unhandled special case #2")
	}

	r1, r2 := findRoots(complex(a, 0.0), complex(b, 0.0), complex(c, 0.0))

	// Pick a positive root for β
	beta := chooseRoot(r1, r2)

	alpha := r * beta

	// Find the control points.
	// wti is p1w
	wti := pi.Add(ti.Multiply(alpha))
	// wte is p4w
	wte := pe.Subtract(te.Multiply(beta))

	//fmt.Printf("wti=%s, wte=%s\r\n", wti.ToString(), wte.ToString())

	// j is the joint point between the two arcs.
	//j := wti.Multiply(beta / (alpha + beta)).Add(wte.Multiply(alpha / (alpha + beta)))

	j := pos.LerpHexFractional(wti, wte, beta/(alpha+beta))
	// tj is the tangent at point j
	//tj := wte.Subtract(wti).Normalize()
	//_, tj, _ := CircularArc{pi, ti, j}.Curve().Sample(1.0)
	// Dumb way #3...
	_, tjp, _ := CircularArc{pe, te.Multiply(-1.0), j}.Curve().Sample(1.0)
	tj := tjp.Multiply(-1.0)

	return []CircularArc{
		CircularArc{pi, ti, j},
		CircularArc{j, tj, pe},
	}
}
