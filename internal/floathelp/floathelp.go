package floathelp

import (
	"math"
)

// CloseEnough tests floats for equality
func CloseEnough(a, b float64) bool {
	// Near zero? Just give up.
	nearZero := 1e-10
	if math.Abs(a) < nearZero && math.Abs(b) < nearZero {
		return true
	}
	epsilon := 1e-4
	return a+epsilon > b && b+epsilon > a
}
