package hexcoord_test

import (
	"fmt"
	"testing"

	"github.com/erinpentecost/hexcoord"
	"github.com/stretchr/testify/assert"
)

type patherImp struct {
	cost map[hexcoord.Hex]int
}

func newPatherImp(size int) patherImp {
	pi := patherImp{
		cost: make(map[hexcoord.Hex]int),
	}

	for h := range concentricMaze(size) {
		// Negative values are impassable
		pi.cost[h] = -1
	}

	return pi
}

func (p patherImp) Cost(a hexcoord.Hex, direction int) int {
	v, ok := p.cost[a.Neighbor(direction)]
	if ok {
		return v
	}
	return 1
}

func (p patherImp) EstimatedCost(a, b hexcoord.Hex) int {
	// This makes the alg perform like Djikstra's alg.
	// Used for testing to help ensure determinism.
	return 0
}

func concentricMaze(maxSize int) <-chan hexcoord.Hex {
	done := make(chan interface{})
	defer close(done)

	mazeGen := make(chan hexcoord.Hex)

	go func() {
		defer close(mazeGen)
		for i := 2; i < maxSize; i = i + 2 {
			opening := i
			cur := 0
			for h := range hexcoord.HexOrigin().RingArea(done, i) {
				cur++
				if opening != cur {
					mazeGen <- h
				}
			}
		}
	}()

	return mazeGen
}

// directPath sets up a test in a map with uniform hex costs.
func directPath(t *testing.T, target hexcoord.Hex) {
	emptyMap := newPatherImp(0)
	path, cost, found := hexcoord.HexOrigin().PathTo(target, emptyMap)

	if found {
		assert.Equal(t, hexcoord.HexOrigin().DistanceTo(target)+1, len(path), fmt.Sprintf("Path to %v (%v away, %v cost) has unexpected length.", target, target.Length(), cost))

		assert.Equal(t, hexcoord.HexOrigin().DistanceTo(target), cost, fmt.Sprintf("Path to %v (%v away) has unexpected cost.", target, target.Length()))

		if len(path) > 0 {
			assert.Equal(t, hexcoord.HexOrigin(), path[0], "First element in path is not the start point.")
			assert.Equal(t, target, path[len(path)-1], "Last element in path is not target point.")
		}
	} else {
		assert.True(t, found, fmt.Sprintf("Can't find path to %v, %v away from source.", target, target.Length()))
	}
}

func TestDirectPaths(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	for i := 1; i < 11; i = i + 2 {
		for h := range hexcoord.HexOrigin().RingArea(done, i) {
			directPath(t, h)
		}
	}
}

// indirectPath sets up a test in a map with different hex costs.
func indirectPath(t *testing.T, target hexcoord.Hex) {
	mazeMap := newPatherImp(target.Length() + 4)
	path, _, found := hexcoord.HexOrigin().PathTo(target, mazeMap)

	if found {
		if len(path) > 0 {
			assert.Equal(t, hexcoord.HexOrigin(), path[0], "First element in path is not the start point.")
			assert.Equal(t, target, path[len(path)-1], "Last element in path is not target point.")
		}
	} else {
		assert.True(t, found, fmt.Sprintf("Can't find path to %v, %v away from source.", target, target.Length()))
	}
}

func TestIndirectPaths(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	for i := 1; i < 11; i = i + 2 {
		for h := range hexcoord.HexOrigin().RingArea(done, i) {
			indirectPath(t, h)
		}
	}
}
