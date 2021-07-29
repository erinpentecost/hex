package pos

import (
	"math"
)

type Area []Hex

// LineArea returns all hexes in a line from point x to point b, inclusive.
// The order of elements is a line as you would expect.
func (h Hex) LineArea(x Hex) Area {
	n := h.DistanceTo(x)
	line := make([]Hex, 0)
	step := 1.0 / math.Max(float64(n), 1.0)
	for i := 0; i <= n; i++ {
		line = append(line, LerpHex(h, x, step*float64(i)))
	}
	return line
}

// HexArea returns the set of hexes that form a larger hex area
// centered around the starting hex and with the given radius.
// The order of elements returned is not set.
// A radius of 0 will return the center hex.
func (h Hex) HexArea(radius int) Area {
	area := make([]Hex, 0)
	for q := -1 * radius; q <= radius; q++ {
		r1 := maxInt(-1*radius, -1*(q+radius))
		r2 := minInt(radius, (-1*q)+radius)

		for r := r1; r <= r2; r++ {
			area = append(area, Hex{
				Q: q,
				R: r,
			})
		}
	}
	return area
}

// RectangleArea returns the set of hexes that form a rectangular
// area from the given hex to another hex representing an opposite corner.
func (h Hex) RectangleArea(opposite Hex) Area {
	area := make([]Hex, 0)

	minR := minInt(h.R, opposite.R)
	maxR := maxInt(h.R, opposite.R)

	minQ := minInt(h.Q, opposite.Q)
	maxQ := maxInt(h.Q, opposite.Q)
	for r := minR; r <= maxR; r++ {
		rOffset := r / 2
		for q := minQ - rOffset; q <= maxQ-rOffset; q++ {
			area = append(area, Hex{
				Q: q,
				R: r,
			})
		}
	}
	return area
}

// RingArea returns a set of hexes that form a ring at the given
// radius centered on the given hex.
// A radius of 0 will return the center hex.
func (h Hex) RingArea(radius int) Area {
	area := make([]Hex, 0)
	if radius == 0 {
		area = append(area, h)
	} else {
		ringH := h.Add(Direction(4).Multiply(radius))
		for i := 0; i < 6; i++ {
			for j := 0; j < radius; j++ {
				area = append(area, ringH)
				ringH = ringH.Neighbor(i)
			}
		}
	}
	return area
}

// SpiralArea returns the set of hexes that form a larger hex area
// centered around the starting hex and with the given radius.
// The order of elements is closest-to-center first.
// If you don't care about the ordering, use HexArea instead.
// A radius of 0 will return the center hex.
func (h Hex) SpiralArea(radius int) Area {
	area := make([]Hex, 0)

	for r := 0; r <= radius; r++ {
		area = append(area, h.RingArea(r)...)
	}

	return area
}

// TriangleArea generates a triangle starting from the given hex
// to points a and b, inclusive.
func (h Hex) TriangleArea(a Hex, b Hex) Area {
	panic("not implemented")
}

// AreaMap applies a function (transform) to each element in
// input (a collection of hexes) and returns a new collection
// with the output.
func AreaMap(input Area, transform func(hex Hex) Hex) Area {

	area := make([]Hex, 0)

	for _, h := range input {
		area = append(area, transform(h))
	}

	return area
}

// AreaFlatMap applies a function (transform) to each element in
// input (a collection of hexes) and returns a new collection
// with the output.
func AreaFlatMap(input Area, transform func(hex Hex) Area) Area {

	area := make([]Hex, 0)

	for _, h := range input {
		area = append(area, transform(h)...)
	}

	return area
}

// combine merges two or more collections of areas, dropping duplicates.
// Set countFilter to 1 to return all hexes that appear at least 1 time.
// This is equivalent to a union of all areas.
// Similarly, set it to n if you only want hexes that appear at least n times.
// This is equivalent to an intersection of all areas.
// The implementation of combine is a little weird. This is to allow elements
// to be sent into the output channel before all inputs are closed.
func combine(countFilter int, areas ...Area) Area {

	combination := make([]Hex, 0)
	// Counts number of times a hex shows up
	// across all input areas.
	seen := make(map[Hex]int)

	// markSeen tracks the hex,
	// and will return how many times it has
	// been marked.
	markSeen := func(h Hex) int {
		// oldVal will be 0 if key does not exist
		oldVal := seen[h]
		seen[h] = oldVal + 1
		return oldVal + 1
	}

	for _, c := range areas {
		for _, h := range c {
			if countFilter == markSeen(h) {
				combination = append(combination, h)
			}
		}
	}

	return combination
}

// AreaUnion returns all hexes in all areas.
// Order is not preserved between areas. Duplicates are removed.
func AreaUnion(areas ...Area) Area {
	return combine(1, areas...)
}

// AreaUnique removes duplicates but retains order.
func AreaUnique(area Area) Area {

	return combine(1, area)
}

// AreaIntersection returns only those hexes that are in all areas.
// Order is not preserved between areas. Duplicates are removed.
func AreaIntersection(areas ...Area) Area {

	areaCount := len(areas)
	return combine(areaCount, areas...)
}

// AreaDifference returns hexes that are only in one area.
// Order is not preserved. Duplicates are removed.
// This won't return any elements until all input channels are closed.
func AreaDifference(areas ...Area) Area {

	difference := make([]Hex, 0)
	// Counts number of times a hex shows up
	// across all input areas.
	seen := make(map[Hex]int)

	// markSeen tracks the hex,
	// and will return how many times it has
	// been marked.
	markSeen := func(h Hex) int {
		// oldVal will be 0 if key does not exist
		oldVal := seen[h]
		seen[h] = oldVal + 1
		return oldVal + 1
	}

	for _, c := range areas {
		for _, h := range c {
			markSeen(h)
		}
	}

	// all inputs are closed at this point,
	// so dump the map into the output channel.
	for k, v := range seen {
		if v == 1 {
			difference = append(difference, k)
		}
	}

	return difference
}

// AreaEqual returns true if all areas have the same exact hexes.
// Ordering is ignored.
func AreaEqual(areas ...Area) bool {
	if len(areas) <= 1 {
		return true
	}

	seen := make(map[Hex]interface{})

	for _, h := range areas[0] {
		seen[h] = nil
	}

	for _, c := range areas[1:] {
		for _, h := range c {
			_, ok := seen[h]
			if !ok {
				return false
			}
		}
	}

	return true
}
