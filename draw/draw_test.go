package draw_test

import (
	"fmt"
	"testing"

	"github.com/erinpentecost/hexcoord/draw"

	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
)

func TestDraw(t *testing.T) {
	dd := draw.DefaultDecorator{}

	img := draw.RenderGrid(500, 0.2, pos.Origin(), dd)

	path, err := draw.Save(img, "testdraw.png")
	assert.NoError(t, err, path)
	fmt.Println(path)
}
