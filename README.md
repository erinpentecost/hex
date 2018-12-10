# hexcoord

[![Go Report Card](https://goreportcard.com/badge/github.com/erinpentecost/hexcoord)](https://goreportcard.com/report/github.com/erinpentecost/hexcoord)
[![Travis CI](https://travis-ci.org/erinpentecost/hexcoord.svg?branch=master)](https://travis-ci.org/erinpentecost/hexcoord.svg?branch=master)
[![GoDoc](https://godoc.org/github.com/erinpentecost/hexcoord?status.svg)](https://godoc.org/github.com/erinpentecost/hexcoord)

![hexcoord](draw/TestDrawLogo.png)

hexcoord is a Go implementation of hexagonal grid math based on amitp's *Hexagonal Grids* articles. This package focuses on hexagonal grid math, including:

* Generating sets of hexes programmatically in common patterns.
* Compositing sets of hexes with unions, intersections, and subtractions.
* A* pathing in a hex grid.
* Super naive drawing package! This isn't performant; it's to help you visualize what's going on.

## References

* [Hexagonal Grids](https://www.redblobgames.com/grids/hexagons)
* [Implementation of Hex Grids](https://www.redblobgames.com/grids/hexagons/implementation.html)
* [Priority Queue](https://golang.org/pkg/container/heap/#example__priorityQueue)
* *Concurrency in Go* by Katherine Cox-Buday