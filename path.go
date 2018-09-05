package hexcoord

// HierarchicalPath-FindingA*
// https://webdocs.cs.ualberta.ca/~mmueller/ps/hpastar.pdf

// Adapt graph code from https://github.com/yourbasic/graph
// My graphs are all super dense, and this one uses hash maps.

// MoveSpeeder determines the move cost between two neighboring
// hex coordinates.
// A cost of 0 indicates impassable terrain.
// Otherwise, higher values indicate more preferable terrain.
type MoveSpeeder interface {
	Cost(a, b Hex) uint8
}
