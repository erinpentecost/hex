// Package fltcmp implements ULP float comparison.
// https://randomascii.wordpress.com/2012/02/25/comparing-floating-point-numbers-2012-edition/
package fltcmp

import (
	"math"
)

// AlmostEqual tells you how close two floats are.
// Make maxUlpsDiff 1 if they need be really, really close.
func AlmostEqual(a, b float64, maxUlpsDiff uint64) bool {
	// Exact case short-circuit.
	if a == b {
		return true
	}

	// Convert to ints
	uA := math.Float64bits(a)
	uB := math.Float64bits(b)

	// Fail when signs are different.
	if signbit64(uA) != signbit64(uB) {
		return false
	}

	// Find difference in ULPs.
	ulpsDiff := diff64(uA, uB)
	return ulpsDiff <= maxUlpsDiff
}

func diff64(a, b uint64) uint64 {
	if a > b {
		return a - b
	}
	return b - a
}

func signbit64(x uint64) bool {
	return x&(1<<63) != 0
}

// AlmostEqual32 tells you how close two float32s are.
// Make maxUlpsDiff 1 if they need be really, really close.
func AlmostEqual32(a, b float32, maxUlpsDiff uint32) bool {
	// Exact case short-circuit.
	if a == b {
		return true
	}

	// Convert to ints
	uA := math.Float32bits(a)
	uB := math.Float32bits(b)

	// Fail when signs are different.
	if signbit32(uA) != signbit32(uB) {
		return false
	}

	// Find difference in ULPs.
	ulpsDiff := diff32(uA, uB)
	return ulpsDiff <= maxUlpsDiff
}

func diff32(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

func signbit32(x uint32) bool {
	return x&(1<<31) != 0
}
