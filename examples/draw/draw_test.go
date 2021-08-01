package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"sync"
	"testing"

	"github.com/erinpentecost/hexcoord/csg"
	"github.com/erinpentecost/hexcoord/pos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDraw(t *testing.T) {
	cc := NewCamera(800, csg.BigHex(pos.Origin(), 2).Subtract(csg.BigHex(pos.Origin(), 1)).Build())
	img := cc.Draw()

	err := Save(img, "testdraw.png")
	assert.NoError(t, err)
}

func createLogoPoints() []pos.Hex {
	h := []pos.Hex{
		{Q: 2, R: -2},
		{Q: 1, R: -1},
		{Q: 0, R: 0},
		{Q: -1, R: 1},
		{Q: 1, R: 0},
		{Q: 1, R: 1},
	}
	ec := []pos.Hex{
		{Q: 2, R: -1},
		{Q: 1, R: -1},
		{Q: 0, R: 0},
		{Q: 0, R: 1},
		{Q: 1, R: 1},
	}
	o := []pos.Hex{
		{Q: 2, R: 0},
		{Q: 2, R: -1},
		{Q: 1, R: -1},
		{Q: 0, R: 0},
		{Q: 0, R: 1},
		{Q: 1, R: 1},
	}
	x := []pos.Hex{
		{Q: 2, R: -1},
		{Q: 1, R: 0},
		{Q: 0, R: 1},
		{Q: 1, R: -1},
		{Q: 1, R: 1},
	}
	r := []pos.Hex{
		{Q: 2, R: 0},
		{Q: 2, R: -1},
		{Q: 1, R: -1},
		{Q: 0, R: 0},
		{Q: -1, R: 1},
	}
	d := []pos.Hex{
		{Q: 2, R: 0},
		{Q: 2, R: -1},
		{Q: 1, R: -1},
		{Q: 0, R: 0},
		{Q: 0, R: 1},
		{Q: 1, R: 1},
		{Q: 3, R: -1},
		{Q: 4, R: -2},
	}

	logo := [][]pos.Hex{
		h,
		ec,
		x,
		ec,
		o,
		o,
		r,
		d,
	}

	taggedPos := make([]pos.Hex, 0)

	for offset, char := range logo {
		for _, h := range char {
			oh := h.Add(pos.Hex{Q: int64(offset * 4), R: 0})
			taggedPos = append(taggedPos, oh)
		}
	}

	return taggedPos
}

func TestDrawLogo(t *testing.T) {

	points := createLogoPoints()

	b, err := json.Marshal(points)
	require.NoError(t, err)

	f, err := os.Create("logo.json")
	require.NoError(t, err)
	defer f.Close()
	f.Write(b)

	cmd := exec.Command("go", "run", "main.go", "-file", "TestDrawLogo.png", "-w", "300")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		writer, err := cmd.StdinPipe()
		if err != nil {
			writer.Close()
		}
		wg.Done()
		writer.Write(b)
	}()
	wg.Wait()

	err = cmd.Run()
	require.NoError(t, err)
}
