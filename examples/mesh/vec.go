package mesh

import "math"

func vecSub(a, b [3]float32) [3]float32 {
	return [3]float32{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}

func vecMagnitude(a [3]float32) float64 {
	ax := float64(a[0])
	ay := float64(a[1])
	az := float64(a[2])
	return math.Sqrt(ax*ax + ay*ay + az*az)
}

func vecCross(a, b [3]float32) [3]float32 {
	return [3]float32{
		a[1]*b[2] - a[2]*b[1],
		a[0]*b[2] - a[2]*b[0],
		a[0]*b[1] - a[1]*b[0],
	}
}
