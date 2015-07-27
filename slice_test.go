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

func TestSliceSuccessorOf(t *testing.T) {
	s := newSlice()
	n := s.successorOf(big.NewInt(int64(8)))
	if n.ID.Cmp(big.NewInt(int64(10))) != 0 {
		t.Errorf("SS failed:%s", n)
	}
	n = s.successorOf(new(big.Int).SetBytes([]byte{0xe}))
	if n != nil {
		t.Errorf("SP failed:%s", n)
	}
}

func TestSlicePredecessorOf(t *testing.T) {
	s := newSlice()
	n := s.predecessorOf(big.NewInt(int64(8)))
	if n.ID.Cmp(big.NewInt(int64(6))) != 0 {
		t.Errorf("SP failed:%s", n)
	}
	n = s.predecessorOf(big.NewInt(int64(2)))
	if n != nil {
		t.Errorf("SS failed:%s", n)
	}
}

func TestPreSucSame(t *testing.T) {

	s := newSlice()
	s.nodes = s.nodes[:3]
	sn := s.successorOf(big.NewInt(int64(4)))
	pn := s.predecessorOf(big.NewInt(int64(4)))

	if sn == pn {
		t.Errorf("SN:%s, PN:%s", sn, pn)
	}

	s.nodes = s.nodes[:2]
	sn = s.successorOf(big.NewInt(int64(4)))
	pn = s.predecessorOf(big.NewInt(int64(4)))

	if sn == pn {
		t.Errorf("SN:%s, PN:%s", sn, pn)
	}
}
