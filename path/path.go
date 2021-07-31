package path

import (
	"container/heap"
	"sync"

	"github.com/erinpentecost/hexcoord/pos"
)

type aStarInfo struct {
	// parent is the hex we moved from to get to this hex.
	// This forms a linked list pointing all the way back to `from`.
	parent pos.Hex
	// cost is the total value from `from` to this hex.
	cost int
}

// unwind walks a field map backwards into a path
func unwind(field map[pos.Hex]aStarInfo, origin pos.Hex, destination pos.Hex) []pos.Hex {
	path := make([]pos.Hex, 0)
	// Unwind to get path
	cur := destination
	curExtras := field[cur]

	// Begin unwind
	for {
		path = append([]pos.Hex{cur}, path...)

		if cur == origin {
			return path
		}

		cur = curExtras.parent
		curExtras = field[cur]
	}
}

// wind walks a field map forwards into a path
func wind(field map[pos.Hex]aStarInfo, origin pos.Hex, destination pos.Hex) []pos.Hex {
	path := make([]pos.Hex, 0)
	// Unwind to get path
	cur := destination
	curExtras := field[cur]

	// Begin unwind
	for {
		path = append(path, cur)

		if cur == origin {
			return path
		}

		cur = curExtras.parent
		curExtras = field[cur]
	}
}

// To finds a near-optimal path to the target hex.
//
// The first element in the path will be the starting hex,
// and the last will be the target hex.
//
// If there is no path, this will be empty.
//
// This is an offline search algorithm; there is no caching.
func To(from pos.Hex, target pos.Hex, pather Pather) (path []pos.Hex) {

	// This is basically two A* searches that run in parallel.
	// One starts at `from`, the other starts at `target`.
	// Once they touch, the paths are stitched together.
	// Both the searches use a priority queue to rank neighbors
	// that need to be searched.

	// This double-headed search is 350840 ns/op on my machine
	// vs 1056354 ns/op for a typical single-threaded A*.

	// Init output variables
	path = make([]pos.Hex, 0)

	// Base case.
	if from == target {
		path = append(path, from)
		return
	}

	// Set up frontier tracker starting at `from`
	fromPaths := make(map[pos.Hex]aStarInfo)
	fromPaths[from] = aStarInfo{
		parent: from,
		cost:   0,
	}

	// Set up frontier tracker starting at `to`
	targetPaths := make(map[pos.Hex]aStarInfo)
	targetPaths[target] = aStarInfo{
		parent: target,
		cost:   0,
	}

	targetMux := sync.Mutex{}
	go func() {
		targetPQ := &priorityQueue{&pqItem{
			Value:    target,
			Priority: 0,
			Index:    0,
		}}
		heap.Init(targetPQ)

		// Cycle through all the neigbors starting at `target`
		for targetPQ.Len() > 0 {
			targetFrontier := (*(heap.Pop(targetPQ).(*pqItem))).Value

			// Look at all neighbors
			for i, next := range targetFrontier.Neighbors() {
				// edgeCost is reversed here
				edgeCost := pather.Cost(next, pos.BoundFacing(i+3))
				// Negative costs are a special case
				if edgeCost < 0 {
					continue
				}

				// Push neighbors we still need to evaluate onto the heap
				targetMux.Lock()
				newCost := targetPaths[targetFrontier].cost + edgeCost
				c, ok := targetPaths[next]
				if !ok || c.cost > newCost {
					// check if main goroutine is done
					if targetPaths == nil {
						targetMux.Unlock()
						return
					}
					// Check if we are extending into main goroutine's seen area
					_, stop := fromPaths[next]
					targetPaths[next] = aStarInfo{
						parent: targetFrontier,
						cost:   newCost,
					}
					targetMux.Unlock()
					if stop {
						return
					}
					heap.Push(targetPQ, &pqItem{
						Value: next,
						// estimatedCost is reversed here
						Priority: newCost + pather.EstimatedCost(from, next),
					})
				} else {
					targetMux.Unlock()
				}
			}
		}
		// no solution if we get to here
	}()

	fromPQ := &priorityQueue{&pqItem{
		Value:    from,
		Priority: 0,
		Index:    0,
	}}
	heap.Init(fromPQ)

	// Cycle through all the neigbors starting at `from`
	for fromPQ.Len() > 0 {
		fromFrontier := (*(heap.Pop(fromPQ).(*pqItem))).Value

		// Quit if the fromFrontier hit a visited node in the targetPaths.
		targetMux.Lock()
		if _, ok := targetPaths[fromFrontier]; ok {
			firstSection := unwind(fromPaths, from, fromFrontier)
			secondSection := wind(targetPaths, target, fromFrontier)

			// join em!
			path = append(firstSection, secondSection[1:]...)

			// let the other A* know to quit
			targetPaths = nil
			targetMux.Unlock()

			// yay we won
			return
		}
		targetMux.Unlock()

		// Look at all neighbors
		for i, next := range fromFrontier.Neighbors() {
			edgeCost := pather.Cost(fromFrontier, i)
			// Negative costs are a special case
			if edgeCost < 0 {
				continue
			}

			// Push neighbors we still need to evaluate onto the heap
			targetMux.Lock()
			newCost := fromPaths[fromFrontier].cost + edgeCost
			c, ok := fromPaths[next]
			if !ok || c.cost > newCost {
				fromPaths[next] = aStarInfo{
					parent: fromFrontier,
					cost:   newCost,
				}
				heap.Push(fromPQ, &pqItem{
					Value:    next,
					Priority: newCost + pather.EstimatedCost(next, target),
				})
			}
			targetMux.Unlock()
		}
	}

	// No solution
	// let the other A* know to quit
	targetMux.Lock()
	targetPaths = nil
	targetMux.Unlock()
	return
}
