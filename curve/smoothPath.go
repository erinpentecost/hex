package curve

import (
	"fmt"
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
	for p := 1; p < len(path)-1; p++ {
		tangents[p] = approximateTangent(path[p-1], path[p], path[p+1])
	}

	// Generate biarcs for each pair of points.
	for i := 0; i < len(path)-1; i++ {
		for _, b := range Biarc(path[i], tangents[i], path[i+1], tangents[i+1], 0.5) {
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

	t := ti.Add(te)

	v := pi.Subtract(pe)

	// This is the line segment case.
	// Start and end points are collinear with
	// the tangents.
	if closeEnough(v.Normalize().DotProduct(t), -1.0) {
		return []CircularArc{
			CircularArc{pi, ti, pe},
		}
	}

	// Now find the positive root for
	// v ⋅ v + 2 β v ⋅ ( r t s + t e ) + 2 r β 2 ( t s ⋅ t e − 1 ) = 0
	// β^2
	a := (ti.DotProduct(te) - 1.0) * 2.0 * r
	// β
	b := v.DotProduct(ti.Multiply(r).Add(te)) * 2.0
	// constant
	c := v.DotProduct(v)

	// Semicircle case
	if closeEnough(a, 0.0) {
		j := pos.LerpHexFractional(pi, pe, 0.5)
		tj := ti.Multiply(-1.0)
		return []CircularArc{
			CircularArc{pi, ti, j},
			CircularArc{j, tj, pe},
		}
	}

	r1, r2 := findRoots(complex(a, 0.0), complex(b, 0.0), complex(c, 0.0))

	// Pick a positive root for β
	var beta float64
	if closeEnough(imag(r1), 0.0) && real(r1) >= 0.0 {
		beta = real(r1)
	} else if closeEnough(imag(r2), 0.0) && real(r2) >= 0.0 {
		beta = real(r2)
	} else {
		panic(fmt.Sprintf("Can't find good roots for %v*β^2+%v*β+%v=0", a, b, c))
	}

	alpha := r * beta

	// Find the control points.
	// wti is p1w
	wti := pi.Add(ti.Multiply(alpha))
	// wte is p4w
	wte := pe.Add(te.Multiply(beta * (-1.0)))

	fmt.Printf("wti=%s, wte=%s", wti.ToString(), wte.ToString())

	// j is the joint point between the two arcs.
	//j := wti.Multiply(beta / (alpha + beta)).Add(wte.Multiply(alpha / (alpha + beta)))
	j := pos.LerpHexFractional(wti, wte, beta/(alpha+beta))
	// tj is the tangent at point j
	tj := wte.Subtract(wti).Normalize()
	//_, tj, _ := CircularArc{pi, ti, j}.Curve().Sample(1.0)
	//if !c1t.AlmostEquals(tj) {
	//	panic("not G0 connected")
	//}

	return []CircularArc{
		CircularArc{pi, ti, j},
		CircularArc{j, tj, pe},
	}
}
