package internal

var (
	RotationMatrixes = [6][4][4]int64{
		{{1, 0, 0, 0}, {0, 1, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}},    // rotate by 0
		{{0, 0, -1, 0}, {-1, 0, 0, 0}, {0, -1, 0, 0}, {0, 0, 0, 1}}, // rotate by 1
		{{0, 1, 0, 0}, {0, 0, 1, 0}, {1, 0, 0, 0}, {0, 0, 0, 1}},    // etc
		{{-1, 0, 0, 0}, {0, -1, 0, 0}, {0, 0, -1, 0}, {0, 0, 0, 1}},
		{{0, 0, 1, 0}, {1, 0, 0, 0}, {0, 1, 0, 0}, {0, 0, 0, 1}},
		{{0, -1, 0, 0}, {0, 0, -1, 0}, {-1, 0, 0, 0}, {0, 0, 0, 1}},
	}
)

func MatrixMultiply(m ...[4][4]int64) [4][4]int64 {
	switch len(m) {
	case 0:
		return [4][4]int64{{1, 0, 0, 0}, {0, 1, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}}
	case 1:
		return m[0]
	}

	out := [4][4]int64{}
	for mx := len(m) - 2; mx >= 0; mx-- {

		x := m[mx]
		y := m[mx+1]

		for i := 0; i < len(x); i++ {
			for j := 0; j < len(y[0]); j++ {
				for k := 0; k < len(y); k++ {
					out[i][j] += x[i][k] * y[k][j]
				}
			}
		}
	}
	return out
}

// BoundFacing maps the whole number set to 0-5.
func BoundFacing(facing int) int {
	d := facing % 6
	if d < 0 {
		d = d + 6
	}
	return d
}

func RotateMatrix(direction int) [4][4]int64 {
	d := BoundFacing(direction)

	return RotationMatrixes[d]
}

// [[1,0,0,tr]
// [0,1,0,tq]
// [0,0,1,0] // this is for s, which is a computed field. ignored.
// [0,0,0,1]] // homogenous coords. ignored.
func TranslateMatrix(r, s int64) [4][4]int64 {
	return [4][4]int64{{1, 0, 0, r}, {0, 1, 0, s}, {0, 0, 1, 0}, {0, 0, 0, 1}}
}
