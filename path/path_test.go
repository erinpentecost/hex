package path_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/erinpentecost/hex"
	"github.com/erinpentecost/hex/area"
	"github.com/erinpentecost/hex/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type patherImp struct {
	cost map[hex.Hex]int
}

func newPatherImp(walls *area.Area) patherImp {
	pi := patherImp{
		cost: make(map[hex.Hex]int),
	}

	for _, h := range walls.Slice() {
		// Negative values are impassable
		pi.cost[h] = -1
	}

	return pi
}

func (p patherImp) Cost(a hex.Hex, direction int) int {
	v, ok := p.cost[a.Neighbor(direction)]
	if ok {
		return v
	}
	return 1
}

func (p patherImp) EstimatedCost(a, b hex.Hex) int {
	// This makes the alg perform like Djikstra's alg.
	// Used for testing to help ensure determinism.
	return 0
}

func ring(center hex.Hex, radius int64) *area.Area {
	return area.BigHex(center, radius).Subtract(area.BigHex(center, radius-1)).Build()
}

func concentricMaze(maxSize int64) *area.Area {
	c := area.NewBuilder()

	for i := int64(2); i < maxSize; i = i + 2 {
		opening := i
		cur := int64(0)
		for _, h := range ring(hex.Origin(), i).Slice() {
			cur++
			if opening != cur {
				c = c.Union(area.NewArea(h))
			}
		}
	}

	return c.Build()
}

// indirectPath sets up a test in a map with different hex costs.
func pathCheck(t *testing.T, target hex.Hex, pather path.Pather) {
	t.Helper()

	path := path.To(hex.Origin(), target, pather)

	require.NotEmpty(t, path, "Can't find path to %v, %v away from source.", target, target.Length())
	require.GreaterOrEqual(t, int64(len(path)), target.DistanceTo(target))
	if len(path) > 0 {
		assert.Equal(t, hex.Origin(), path[0], "First element %s in path is not the start point %s.", path[0], hex.Origin())
		assert.Equal(t, target, path[len(path)-1], "Last element %s in path is not target point %s.", path[len(path)-1], target)
	}

	// make sure path is contiguous with no loops
	seen := make(map[hex.Hex]interface{})
	last := hex.Origin()
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
			require.EqualValues(t, 1, last.DistanceTo(p), "Path is not contiguous between idx=%d pos=%s and idx=%d pos=%s.\npath=%s", i-1, last.String(), i, p.String(), sb.String())
		}
		last = p
	}
}

func TestDirectPaths(t *testing.T) {
	for i := int64(1); i < 11; i = i + 2 {
		for _, h := range ring(hex.Origin(), i).Slice() {
			t.Run(fmt.Sprintf("to-%s", h.String()), func(t *testing.T) {
				pathCheck(t, h, newPatherImp(area.NewArea()))
			})
		}
	}
}

func TestIndirectPaths(t *testing.T) {
	for i := int64(1); i < 11; i = i + 2 {
		for _, h := range ring(hex.Origin(), i).Slice() {
			t.Run(fmt.Sprintf("to-%s", h.String()), func(t *testing.T) {
				pathCheck(t, h, newPatherImp(concentricMaze(h.Length()+4)))
			})
		}
	}
}

func TestNoPath(t *testing.T) {
	t.Parallel()

	pather := newPatherImp(ring(hex.Origin(), 5))

	foundPath := path.To(hex.Origin(), hex.Hex{Q: 100, R: 100}, pather)
	require.Empty(t, foundPath)
}

func BenchmarkDirectPath(b *testing.B) {
	var foundPath []hex.Hex
	target := hex.Hex{Q: 10, R: 10}
	pather := newPatherImp(area.NewArea())
	for i := 0; i < b.N; i++ {
		foundPath = path.To(hex.Origin(), target, pather)
	}
	assert.Equal(b, hex.Origin().DistanceTo(target)+1, len(foundPath))
}
