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
}
