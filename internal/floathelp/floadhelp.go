package floathelp

import "github.com/erinpentecost/fltcmp"

// CloseEnough tests floats for equality
func CloseEnough(a, b float64) bool {
	return fltcmp.AlmostEqual(a, b, 15)
}
