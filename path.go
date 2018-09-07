package hexcoord

import (
	"container/heap"
)

type pqItem struct {
	value    Hex
	priority int
	index    int
}

// priorityQueue from https://golang.org/pkg/container/heap/#example__priorityQueue
type priorityQueue []*pqItem

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*pqItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func (pq *priorityQueue) update(item *pqItem, value Hex, priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}

type aStarInfo struct {
	parent Hex
	cost   int
}

// Pather contains domain knowledge for finding a path.
type Pather interface {
	// Cost indicates the move cost between a hex and one
	// of its neighbors. Higher values are less desirable.
	Cost(a Hex, direction int) int

	// EstimatedCost returns the estimated cost between
	// two hexes that are not necessarily neighbors.
	//
	EstimatedCost(a, b Hex) int
}

// PathTo finds a near-optimal path to the target hex.
func (h Hex) PathTo(target Hex, pather Pather) (path []Hex, cost int, found bool) {
	// Init output variables
	path = make([]Hex, 0)
	cost = 0
	found = false

	if h == target {
		found = true
		return
	}

	// Init supporting data structures.
	pq := make(priorityQueue, 1)
	extras := make(map[Hex]aStarInfo)

	// Prime with start node.
	pq[0] = &pqItem{
		value:    h,
		priority: 0,
		index:    0,
	}
	extras[h] = aStarInfo{
		parent: h,
		cost:   0,
	}

	// Begin A*
	for pq.Len() > 0 {
		currentHeapItem := *(pq.Pop().(*pqItem))
		current := currentHeapItem.value

		// Quit if we found it
		if current == target {
			found = true
			break
		}

		// Look at all neigbors
		for i, next := range h.Neighbors() {
			newCost := extras[current].cost + pather.Cost(current, i)
			c, ok := extras[next]
			if !ok {
				extras[next] = aStarInfo{
					parent: current,
					cost:   newCost,
				}
				pq.Push(&pqItem{
					value:    next,
					priority: newCost + pather.EstimatedCost(next, target),
				})
			} else if c.cost > newCost {
				// Since we ran into a point we already saw,
				// lets check if this is a better path to it
				// than the previous one and swap if it is.
				extras[next] = aStarInfo{
					parent: current,
					cost:   newCost,
				}
				// Need to also update it in the heap, probably...
			}
		}
	}

	// Quit if the target is not in the found set.
	if !found {
		return
	}

	// Unwind to get path
	cur := target
	curExtras, _ := extras[cur]

	// Begin unwind
	for {
		path = append(path, cur)
		cost = cost + curExtras.cost

		if cur == h {
			return
		}

		cur = curExtras.parent
		curExtras, _ = extras[cur]
	}
}
