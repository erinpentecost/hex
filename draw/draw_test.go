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
	cc := draw.NewCamera(600, 500, 0.2, pos.Hex{Q: -1, R: -1})

	img := cc.Render(dd)

	path, err := draw.Save(img, "testdraw.png")
	assert.NoError(t, err, path)
	fmt.Println(path)
}
