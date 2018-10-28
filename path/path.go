package path

import (
	"container/heap"

	"github.com/erinpentecost/hexcoord/pos"
)

type pqItem struct {
	value    pos.Hex
	priority int
	index    int
}

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

type aStarInfo struct {
	parent pos.Hex
	cost   int
}

// Pather contains domain knowledge for finding a path.
type Pather interface {
	// Cost indicates the move cost between a hex and one
	// of its neighbors. Higher values are less desirable.
	// Negative costs are treated as impassable.
	Cost(a pos.Hex, direction int) int

	// EstimatedCost returns the estimated cost between
	// two hexes that are not necessarily neighbors.
	// Negative costs are treated as impassable.
	EstimatedCost(a, b pos.Hex) int
}

// To finds a near-optimal path to the target hex.
// The first element in the path will be the starting hex,
// and the last will be the target hex.
func To(from pos.Hex, target pos.Hex, pather Pather) (path []pos.Hex, cost int, found bool) {
	// Init output variables
	path = make([]pos.Hex, 0)
	cost = 0
	found = false

	if from == target {
		found = true
		return
	}

	// Init supporting data structures.
	pq := &priorityQueue{&pqItem{
		value:    from,
		priority: 0,
		index:    0,
	}}
	heap.Init(pq)

	extras := make(map[pos.Hex]aStarInfo)
	extras[from] = aStarInfo{
		parent: from,
		cost:   0,
	}

	// Begin A*
	for pq.Len() > 0 {
		currentHeapItem := *(heap.Pop(pq).(*pqItem))
		current := currentHeapItem.value

		// Quit if we found it
		if current == target {
			found = true
			break
		}

		// Look at all neigbors
		for i, next := range current.Neighbors() {
			edgeCost := pather.Cost(current, i)
			// Negative costs are a special case
			if edgeCost < 0 {
				continue
			}
			newCost := extras[current].cost + edgeCost
			c, ok := extras[next]
			if !ok || c.cost > newCost {
				extras[next] = aStarInfo{
					parent: current,
					cost:   newCost,
				}
				heap.Push(pq, &pqItem{
					value:    next,
					priority: newCost + pather.EstimatedCost(next, target),
				})
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
	cost = curExtras.cost

	// Begin unwind
	for {
		path = append([]pos.Hex{cur}, path...)

		if cur == from {
			return
		}

		cur = curExtras.parent
		curExtras, _ = extras[cur]
	}
}
