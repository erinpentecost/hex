package hex

import "sort"

var (
	_ sort.Interface = (Sort)(nil)
)

type Sort []Hex

func (s Sort) Len() int {
	return len(s)
}

func (s Sort) Swap(i int, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Sort) Less(i int, j int) bool {
	if s[i].Q > s[j].Q {
		return true
	}
	return s[i].R > s[j].R
}
