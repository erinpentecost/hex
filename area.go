package hexcoord

import (
	"math"
	"sync"
)

// Area wraps up hexes into an Area.
func Area(h ...Hex) <-chan Hex {
	hgen := make(chan Hex)
	go func() {
		defer close(hgen)
		for _, c := range h {
			hgen <- c
		}
	}()
	return hgen
}

// AreaToSlice iterates over a chan of hexes and stores them in
// in a single slice.
func AreaToSlice(area <-chan Hex) []Hex {
	hslice := make([]Hex, 0)
	for h := range area {
		hslice = append(hslice, h)
	}
	return hslice
}

// LineArea returns all hexes in a line from point x to point b, inclusive.
// The order of elements is a line as you would expect.
func (h Hex) LineArea(done <-chan interface{}, x Hex) <-chan Hex {
	hgen := make(chan Hex)
	go func() {
		defer close(hgen)
		n := h.DistanceTo(x)
		step := 1.0 / math.Max(float64(n), 1.0)
		for i := 0; i <= n; i++ {
			select {
			case <-done:
				return
			case hgen <- lerpHex(h, x, step*float64(i)):
			}
		}
	}()
	return hgen
}

// HexArea returns the set of hexes that form a larger hex area
// centered around the starting hex and with the given radius.
// The order of elements returned is not set.
// A radius of 0 will return the center hex.
func (h Hex) HexArea(done <-chan interface{}, radius int) <-chan Hex {
	hgen := make(chan Hex)
	go func() {
		defer close(hgen)
		for q := -1 * radius; q <= radius; q++ {
			r1 := maxInt(-1*radius, -1*(q+radius))
			r2 := minInt(radius, (-1*q)+radius)

			for r := r1; r <= r2; r++ {
				select {
				case <-done:
					return
				case hgen <- Hex{
					Q: q,
					R: r,
				}:
				}
			}
		}
	}()
	return hgen
}

// RingArea returns a set of hexes that form a ring at the given
// radius centered on the given hex.
// A radius of 0 will return the center hex.
func (h Hex) RingArea(done <-chan interface{}, radius int) <-chan Hex {
	hgen := make(chan Hex)
	go func() {
		defer close(hgen)
		if radius == 0 {
			hgen <- h
		} else {
			ringH := h.Add(HexDirection(4).Multiply(radius))
			for i := 0; i < 6; i++ {
				for j := 0; j < radius; j++ {
					select {
					case <-done:
						return
					case hgen <- ringH:
					}
					ringH = ringH.Neighbor(i)
				}
			}
		}
	}()
	return hgen
}

// SpiralArea returns the set of hexes that form a larger hex area
// centered around the starting hex and with the given radius.
// The order of elements is closest-to-center first.
// If you don't care about the ordering, use HexArea instead.
// A radius of 0 will return the center hex.
func (h Hex) SpiralArea(done <-chan interface{}, radius int) <-chan Hex {
	hgen := make(chan Hex)
	go func() {
		defer close(hgen)
		for r := 0; r <= radius; r++ {
			for ring := range h.RingArea(done, r) {
				select {
				case <-done:
					return
				case hgen <- ring:
				}
			}
		}
	}()
	return hgen
}

// AreaMap applies a function (transform) to each element in
// input (a collection of hexes) and returns a new collection
// with the output.
func AreaMap(
	done <-chan interface{},
	input <-chan Hex,
	transform func(hex Hex) Hex) <-chan Hex {

	hgen := make(chan Hex)

	go func() {
		defer close(hgen)
		for h := range input {
			select {
			case <-done:
				return
			case hgen <- transform(h):
			}
		}
	}()
	return hgen
}

// AreaFlatMap applies a function (transform) to each element in
// input (a collection of hexes) and returns a new collection
// with the output.
func AreaFlatMap(
	done <-chan interface{},
	input <-chan Hex,
	transform func(done <-chan interface{}, hex Hex) <-chan Hex) <-chan Hex {

	agen := make(chan (<-chan Hex))

	// Convert each input hex into a chan of chans
	// by applying transform()
	go func() {
		defer close(agen)
		for h := range input {
			select {
			case <-done:
				return
			case agen <- (transform(done, h)):
			}
		}
	}()

	return AreaSum(done, agen)
}

// AreaSum flattens input areas into a single area using the bridge
// pattern. This does not remove duplicates, but does preserve order.
func AreaSum(done <-chan interface{}, areas <-chan <-chan Hex) <-chan Hex {
	flatStream := make(chan Hex)

	go func() {
		defer close(flatStream)

		for {
			// Pop off area
			var area <-chan Hex
			select {
			case <-done:
				return
			case foundArea, ok := <-areas:
				if !ok {
					return
				}
				area = foundArea
			}

			// Pop off hexes from area
		loop:
			for {
				select {
				case <-done:
					break loop
				case foundHex, ok := <-area:
					if !ok {
						break loop
					}
					flatStream <- foundHex
				}
			}
		}
	}()

	return flatStream
}

// combine merges two or more collections of areas, dropping duplicates.
// Set countFilter to 1 to return all hexes that appear at least 1 time.
// This is equivalent to a union of all areas.
// Similarly, set it to n if you only want hexes that appear at least n times.
// This is equivalent to an intersection of all areas.
// The implementation of combine is a little weird. This is to allow elements
// to be sent into the output channel before all inputs are closed.
func combine(
	done <-chan interface{},
	countFilter int,
	areas ...<-chan Hex) <-chan Hex {

	var wg sync.WaitGroup
	combination := make(chan Hex)
	// Counts number of times a hex shows up
	// across all input areas.
	seen := make(map[Hex]int)
	var seenLock sync.Mutex

	// markSeen tracks the hex,
	// and will return how many times it has
	// been marked.
	markSeen := func(h Hex) int {
		defer seenLock.Unlock()
		seenLock.Lock()
		// oldVal will be 0 if key does not exist
		oldVal, _ := seen[h]
		seen[h] = oldVal + 1
		return oldVal + 1
	}

	multiplex := func(c <-chan Hex) {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			case h, ok := <-c:
				if !ok {
					return
				} else if countFilter == markSeen(h) {
					combination <- h
				}
			}
		}
	}

	wg.Add(len(areas))
	for _, c := range areas {
		go multiplex(c)
	}

	go func() {
		defer close(combination)
		wg.Wait()
	}()

	return combination
}

// AreaUnion returns all hexes in all areas.
// Order is not preserved. Duplicates are removed.
func AreaUnion(
	done <-chan interface{},
	areas ...<-chan Hex) <-chan Hex {

	return combine(done, 1, areas...)
}

// AreaIntersection returns only those hexes that are in all areas.
// Order is not preserved. Duplicates are removed.
func AreaIntersection(
	done <-chan interface{},
	areas ...<-chan Hex) <-chan Hex {

	areaCount := len(areas)
	return combine(done, areaCount, areas...)
}

// AreaDifference returns hexes that are only in one area.
// Order is not preserved. Duplicates are removed.
// This won't return any elements until all input channels are closed.
func AreaDifference(
	done <-chan interface{},
	areas ...<-chan Hex) <-chan Hex {

	var wg sync.WaitGroup
	difference := make(chan Hex)
	// Counts number of times a hex shows up
	// across all input areas.
	seen := make(map[Hex]int)
	var seenLock sync.Mutex

	// markSeen tracks the hex,
	// and will return how many times it has
	// been marked.
	markSeen := func(h Hex) int {
		defer seenLock.Unlock()
		seenLock.Lock()
		// oldVal will be 0 if key does not exist
		oldVal, _ := seen[h]
		seen[h] = oldVal + 1
		return oldVal + 1
	}

	multiplex := func(c <-chan Hex) {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			case h, ok := <-c:
				if !ok {
					return
				}
				markSeen(h)
			}
		}
	}

	wg.Add(len(areas))
	for _, c := range areas {
		go multiplex(c)
	}

	go func() {
		defer close(difference)
		wg.Wait()
		// all inputs are closed at this point,
		// so dump the map into the output channel.
		for k, v := range seen {
			if v == 1 {
				select {
				case <-done:
					return
				case difference <- k:
				}
			}
		}
	}()

	return difference
}

// AreaEqual returns true if all areas have the same exact hexes.
// Ordering is ignored.
func AreaEqual(areas ...<-chan Hex) bool {
	done := make(chan interface{})
	defer close(done)

	// Areas are not equal if differences has at least
	// one element
	differences := AreaDifference(done, areas...)
	_, ok := <-differences
	return !ok
}
