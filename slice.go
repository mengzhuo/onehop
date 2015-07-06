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

func (s *Slice) GetUnit(id *big.Int) *Unit {

	i := sort.Search(len(s.units),
		func(i int) bool {
			return s.units[i].Min.Cmp(id) < 0 && s.units[i].Max.Cmp(id) >= 0
		})

	if i < len(s.units) {
		return s.units[i]
	}

	return nil
}
