# hexcoord

[![Go Report Card](https://goreportcard.com/badge/github.com/erinpentecost/deterbus)](https://goreportcard.com/report/github.com/erinpentecost/hexcoord)
[![Travis CI](https://travis-ci.org/erinpentecost/hexcoord.svg?branch=master)](https://travis-ci.org/erinpentecost/hexcoord.svg?branch=master)

hexcoord is a Go implementation of hexagonal grid math based on amitp's *Hexagonal Grids* articles. This package focuses on hexagonal grid math. It doesn't concern itself with the rendering or storage of hexes or hex maps.

## [hex.go](../master/hex.go)

hex.go contains Hex, the base coordinate type defined in this package, along with a bunch of functions for transforming Hexes.

## [area.go](../master/area.go)

area.go contains functions for creating and manipulating Hex [pipelines](https://blog.golang.org/pipelines). Use the functions here to procedurally generate shapes like lines, rings, and so on.

## [path.go](../master/path.go)

path.go contains an implementation of A* that works on hexagonal maps.

## References

* [Hexagonal Grids](https://www.redblobgames.com/grids/hexagons)
* [Implementation of Hex Grids](https://www.redblobgames.com/grids/hexagons/implementation.html)
* [Priority Queue](https://golang.org/pkg/container/heap/#example__priorityQueue)