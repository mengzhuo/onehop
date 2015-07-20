package onehop

import (
	"math/big"
	"testing"
)

func newSlice() *Slice {
	s := NewRoute(2)
	AddRouteNode(s)
	return s.slices[0]

}

func TestSliceAdd(t *testing.T) {
	s := newSlice()
	n := s.Get(big.NewInt(int64(8)))
	if n == nil {
		t.Errorf("Get error")
	}
	if s.Leader != n {
		t.Errorf("Wrong leader")

	}
}

func TestSliceDelete(t *testing.T) {
	s := newSlice()
	ok := s.Delete(big.NewInt(int64(8)))
	if !ok {
		t.Errorf("delete failed")
	}

	if s.Len() != 6 {
		t.Errorf("delete failed %d", s.Len())
	}

	if s.Leader.ID.Cmp(big.NewInt(int64(10))) != 0 {
		t.Errorf("Delete failed")
	}

	ok = s.Delete(big.NewInt(int64(2)))
	if !ok {
		t.Errorf("delete failed")
	}
	if s.Len() != 5 {
		t.Errorf("delete failed %d", s.Len())
	}
	if s.Leader.ID.Cmp(big.NewInt(int64(10))) != 0 {
		t.Errorf("Delete failed")
	}
}
