package csg

import (
	"sync"

	"github.com/erinpentecost/hexcoord/pos"
)

// Bounding is the return type for CheckBounding.
type Bounding byte

const (
	// Undefined means one or both areas are empty.
	Undefined Bounding = iota
	// Distinct means areas a and b have no hexes in common.
	Distinct
	// Overlap means a and b have at least one hex in common,
	//
	// a has at least one hex not in b, and
	//
	// b has at least one hex not in a.
	Overlap
	// Contains means all the hexes in b are also in a.
	Contains
	// Contains means all the hexes in a are also in b.
	ContainedBy
	// Equals means areas a and b are the same.
	Equals
)

// CheckBounding returns the overlap relationship between a and b.
func (a *Area) CheckBounding(b *Area) Bounding {
	// use bounding boxes to determine if there is any overlap
	// if there is, do a finer check. otherwise return Distinct.

	if len(a.hexes) == 0 || len(b.hexes) == 0 {
		return Undefined
	}

	if a.mightOverlap(b) {
		return a.checkFineBounding(b)
	}
	return Distinct
}

// mightOverlap returns true if the bounding boxes of a and b
// might overlap.
func (a *Area) mightOverlap(b *Area) bool {
	if len(a.hexes) == 0 || len(b.hexes) == 0 {
		return false
	}
	a.ensureBounds()
	b.ensureBounds()
	qOverlap := contains(a.minQ, a.maxQ, b.minQ) ||
		contains(a.minQ, a.maxQ, b.maxQ) ||
		(b.minQ <= a.minQ && b.maxQ >= a.maxQ)

	rOverlap := contains(a.minR, a.maxR, b.minR) ||
		contains(a.minR, a.maxR, b.maxR) ||
		(b.minR <= a.minR && b.maxR >= a.maxR)

	return qOverlap && rOverlap
}

func contains(min, max, test int64) bool {
	return min <= test && max >= test
}

func (a *Area) checkFineBounding(b *Area) Bounding {
	if len(a.hexes) == 0 || len(b.hexes) == 0 {
		return Undefined
	}

	contains := true
	containedBy := true
	overlap := false

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for k := range a.hexes {
			if _, ok := b.hexes[k]; ok {
				overlap = true
			} else {
				containedBy = false
			}
		}
	}()

	for k := range b.hexes {
		if _, ok := a.hexes[k]; ok {
			overlap = true
		} else {
			contains = false
		}
	}

	wg.Wait()

	if !overlap {
		return Distinct
	}
	if contains && containedBy {
		return Equals
	}
	if contains {
		return Contains
	}
	if containedBy {
		return ContainedBy
	}

	return Overlap
}

type boundsFinder struct {
	empty bool
	minR  int64
	maxR  int64
	minQ  int64
	maxQ  int64
}

func (b *boundsFinder) visit(p *pos.Hex) {
	if b.empty {
		b.minR = p.R
		b.maxR = p.R
		b.minQ = p.Q
		b.maxQ = p.Q
		return
	}
	b.minR = minInt(b.minR, p.R)
	b.maxR = maxInt(b.maxR, p.R)

	b.minQ = minInt(b.minQ, p.Q)
	b.maxQ = maxInt(b.maxQ, p.Q)
}

func (b *boundsFinder) applyTo(a *Area) *Area {
	if b.empty {
		a.minR = 0
		a.maxR = 0
		a.minQ = 0
		a.maxQ = 0
		a.boundsClean = false
		return a
	}
	a.minR = b.minR
	a.maxR = b.maxR
	a.minQ = b.minQ
	a.maxQ = b.maxQ
	a.boundsClean = true
	return a
}
