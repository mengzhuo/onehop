package onehop

import (
	"testing"
	"time"
)

func newSlice() *Slice {

	s := NewSlice(ZERO_ID, FULL_ID)

	for id := 0; id < 32; id += 2 {

		s.Add(&Node{BytesToId([]byte{byte(id)}),
			nil, time.Now()})
	}

	return s
}

func TestSliceAdd(t *testing.T) {
	s := newSlice()
	n := s.Get(BytesToId([]byte{16}))
	if n == nil {
		t.Errorf("Get error %s", n)
	}
	if s.Leader() != n {
		t.Errorf("Wrong leader")

	}
}

func TestSliceDelete(t *testing.T) {
	s := newSlice()
	ok := s.Delete(BytesToId([]byte{16}))
	if !ok {
		t.Errorf("delete failed")
	}

	if s.Len() != 15 {
		t.Errorf("delete failed %d", s.Len())
	}

	if s.Leader().ID != BytesToId([]byte{14}) {
		t.Error("Delete failed", s.Leader().ID)
	}

	ok = s.Delete(BytesToId([]byte{30}))
	if !ok {
		t.Errorf("delete failed")
	}
	if s.Len() != 14 {
		t.Errorf("delete failed %d", s.Len())
	}
	if s.Leader().ID != BytesToId([]byte{14}) {
		t.Error("Delete failed", s.Leader())
	}
}

func TestSliceSuccessorOf(t *testing.T) {
	s := newSlice()
	n := s.successorOf(BytesToId([]byte{8}))
	if n.ID != BytesToId([]byte{10}) {
		t.Errorf("SS failed:%s", n)
	}
	n = s.successorOf(BytesToId([]byte{32}))
	if n != nil {
		t.Errorf("SP failed:%s", n)
	}
}

func TestSlicePredecessorOf(t *testing.T) {
	s := newSlice()
	n := s.predecessorOf(BytesToId([]byte{8}))
	if n.ID != BytesToId([]byte{6}) {
		t.Errorf("SP failed:%s", n)
	}
	n = s.predecessorOf(BytesToId([]byte{0}))
	if n != nil {
		t.Errorf("SS failed:%s", n)
	}
}

func TestPreSucSame(t *testing.T) {

	s := newSlice()
	s.Nodes = s.Nodes[:3]
	sn := s.successorOf(BytesToId([]byte{4}))
	pn := s.predecessorOf(BytesToId([]byte{4}))

	if sn == pn {
		t.Errorf("SN:%s, PN:%s", sn, pn)
	}

	s.Nodes = s.Nodes[:2]
	sn = s.successorOf(BytesToId([]byte{4}))
	pn = s.predecessorOf(BytesToId([]byte{4}))

	if sn == pn {
		t.Errorf("SN:%s, PN:%s", sn, pn)
	}
}
