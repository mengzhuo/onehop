package onehop

import "math/big"

type Slice struct {
	Leader *Node
	Min    *big.Int
	Max    *big.Int
	// Units should be preallocated and sorted
	units []*Unit
}

func (s *Slice) updateLeader() {

	for i := len(s.units) / 2; i >= 0; i-- {
		if leader := s.units[i].Leader; leader != nil {
			s.Leader = leader
			return
		}

		j := len(s.units) - i - 1
		if leader := s.units[j].Leader; leader != nil {
			s.Leader = leader
			return
		}
	}
	s.Leader = nil
}

func (s *Slice) Len() int {
	i := 0
	for _, u := range s.units {
		i += u.Len()
	}
	return i
}
