package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"sync"
	"testing"

	"github.com/erinpentecost/hex"
	"github.com/erinpentecost/hex/area"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDraw(t *testing.T) {
	cc := NewCamera(800,
		area.BigHex(hex.Origin(), 2).Subtract(area.BigHex(hex.Origin(), 1)).Build(),
		func(h hex.Hex) string { return "" },
	)
	img := cc.Draw()

	fileHandle, err := os.Create("testdraw.png")
	require.NoError(t, err)
	err = Save(img, fileHandle)
	assert.NoError(t, err)
}

func createLogoPoints() []hex.Hex {
	h := []hex.Hex{
		{Q: 2, R: -2},
		{Q: 1, R: -1},
		{Q: 0, R: 0},
		{Q: -1, R: 1},
		{Q: 1, R: 0},
		{Q: 1, R: 1},
	}
	e := []hex.Hex{
		{Q: 2, R: -1},
		{Q: 1, R: -1},
		{Q: 0, R: 0},
		{Q: 0, R: 1},
		{Q: 1, R: 1},
	}
	x := []hex.Hex{
		{Q: 2, R: -1},
		{Q: 1, R: 0},
		{Q: 0, R: 1},
		{Q: 1, R: -1},
		{Q: 1, R: 1},
	}

	logo := [][]hex.Hex{
		h,
		e,
		x,
	}

	taggedPos := make([]hex.Hex, 0)

	for offset, char := range logo {
		for _, h := range char {
			oh := h.Add(hex.Hex{Q: int64(offset * 4), R: 0})
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
