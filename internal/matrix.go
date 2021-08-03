package internal

var (
	RotationMatrixes = [6][4][4]int64{
		{{1, 0, 0, 0}, {0, 1, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}},    // rotate by 0 (identity)
		{{0, 0, -1, 0}, {-1, 0, 0, 0}, {0, -1, 0, 0}, {0, 0, 0, 1}}, // rotate by 1
		{{0, 1, 0, 0}, {0, 0, 1, 0}, {1, 0, 0, 0}, {0, 0, 0, 1}},    // etc
		{{-1, 0, 0, 0}, {0, -1, 0, 0}, {0, 0, -1, 0}, {0, 0, 0, 1}},
		{{0, 0, 1, 0}, {1, 0, 0, 0}, {0, 1, 0, 0}, {0, 0, 0, 1}},
		{{0, -1, 0, 0}, {0, 0, -1, 0}, {-1, 0, 0, 0}, {0, 0, 0, 1}},
	}
)

func MatrixMultiply(m ...[4][4]int64) [4][4]int64 {
	out := [4][4]int64{{1, 0, 0, 0}, {0, 1, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}}

	for mx := 0; mx < len(m); mx++ {

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
func TranslateMatrix(q, r, s int64) [4][4]int64 {
	// TODO: do I need to use S? probably yes if I want to combine matrices
	return [4][4]int64{{1, 0, 0, q}, {0, 1, 0, r}, {0, 0, 1, s}, {0, 0, 0, 1}}
}
