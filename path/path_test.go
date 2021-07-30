package path_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/erinpentecost/hexcoord/csg"
	"github.com/erinpentecost/hexcoord/path"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type patherImp struct {
	cost map[pos.Hex]int
}

func newPatherImp(size int) patherImp {
	pi := patherImp{
		cost: make(map[pos.Hex]int),
	}

	for h := range concentricMaze(size) {
		// Negative values are impassable
		pi.cost[h] = -1
	}

	return pi
}

func (p patherImp) Cost(a pos.Hex, direction int) int {
	v, ok := p.cost[a.Neighbor(direction)]
	if ok {
		return v
	}
	return 1
}

func (p patherImp) EstimatedCost(a, b pos.Hex) int {
	// This makes the alg perform like Djikstra's alg.
	// Used for testing to help ensure determinism.
	return 0
}

func concentricMaze(maxSize int) csg.Area {
	c := csg.NewBuilder()

	for i := 2; i < maxSize; i = i + 2 {
		opening := i
		cur := 0
		for h := range csg.Ring(pos.Origin(), i).Build() {
			cur++
			if opening != cur {
				c = c.Union(csg.NewArea(h))
			}
		}
	}

	return c.Build()
}

// indirectPath sets up a test in a map with different hex costs.
func pathCheck(t *testing.T, target pos.Hex, pather path.Pather) {
	t.Helper()

	path := path.To(pos.Origin(), target, pather)

	require.NotEmpty(t, path, "Can't find path to %v, %v away from source.", target, target.Length())
	require.GreaterOrEqual(t, len(path), target.DistanceTo(target))
	if len(path) > 0 {
		assert.Equal(t, pos.Origin(), path[0], "First element %s in path is not the start point %s.", path[0], pos.Origin())
		assert.Equal(t, target, path[len(path)-1], "Last element %s in path is not target point %s.", path[len(path)-1], target)
	}

	// make sure path is contiguous with no loops
	seen := make(map[pos.Hex]interface{})
	last := pos.Origin()
	sb := strings.Builder{}
	for _, p := range path {
		sb.WriteString(p.String())
		sb.WriteString(">")
	}
	for i, p := range path {
		// assert no loops
		if _, ok := seen[p]; ok {
			require.FailNow(t, "Oh no!", "Position is duplicated in found path. index=%d pos=%s.\npath=%s", i, p.String(), sb.String())
		}
		seen[p] = nil
		// assert contiguous
		if i != 0 {
			require.Equal(t, 1, last.DistanceTo(p), "Path is not contiguous between idx=%d pos=%s and idx=%d pos=%s.\npath=%s", i-1, last.String(), i, p.String(), sb.String())
		}
		last = p
	}
}

func TestDirectPaths(t *testing.T) {
	t.Parallel()
	for i := 1; i < 11; i = i + 2 {
		for h := range csg.Ring(pos.Origin(), i) {
			t.Run(fmt.Sprintf("to-%s", h.String()), func(t *testing.T) {
				t.Parallel()
				pathCheck(t, h, newPatherImp(0))
			})
		}
	}
}

func TestIndirectPaths(t *testing.T) {
	t.Parallel()
	for i := 1; i < 11; i = i + 2 {
		for h := range csg.Ring(pos.Origin(), i) {
			t.Run(fmt.Sprintf("to-%s", h.String()), func(t *testing.T) {
				t.Parallel()
				pathCheck(t, h, newPatherImp(h.Length()+4))
			})
		}
	}
}

func BenchmarkDirectPath(b *testing.B) {
	var foundPath []pos.Hex
	target := pos.Hex{Q: 10, R: 10}
	pather := newPatherImp(0)
	for i := 0; i < b.N; i++ {
		foundPath = path.To(pos.Origin(), target, pather)
	}
	assert.Equal(b, pos.Origin().DistanceTo(target)+1, len(foundPath))
}
