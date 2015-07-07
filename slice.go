package onehop

import (
	"math/big"
	"sort"
)

type Slice struct {
	Leader *Node
	Min    *big.Int
	Max    *big.Int
	// Units should be preallocated and sorted
	units []*Unit
}

func (s *Slice) GetUnitIndex(id *big.Int) int {

	if id.Cmp(s.Min) <= 0 || id.Cmp(s.Max) > 0 {
		// Not in our slice
		return len(s.units)
	}

	i := sort.Search(len(s.units),
		func(i int) bool {
			return s.units[i].Min.Cmp(id) >= 0
		})

	if i < len(s.units) && i > 0 {
		i -= 1
	}

	if i == 0 {
		return 0
	}

	return i
}

func (s *Slice) Len() int {
	i := 0
	for _, u := range s.units {
		i += u.Len()
	}
	return i
}
