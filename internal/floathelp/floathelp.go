package floathelp

import (
	"math"

	"github.com/erinpentecost/fltcmp"
)

// CloseEnough tests floats for equality
func CloseEnough(a, b float64) bool {
	// Near zero? Just give up.
	nearZero := 1e-10
	if math.Abs(a) < nearZero && math.Abs(b) < nearZero {
		return true
	}

	return fltcmp.AlmostEqual(a, b, 100)
}
