package onehop

import (
	"math/big"
	"testing"
)

var magic7 = big.NewInt(int64(7))

func NewTestUnit() *Unit {
	min := big.NewInt(int64(1))
	max := big.NewInt(int64(16))

	u := NewUnit(min, max)
	return u
}

func TestNewUnit(t *testing.T) {

	u := NewTestUnit()
	if u.Max.Cmp(big.NewInt(int64(16))) != 0 {
		t.Error("u.Max not vaild")
	}
}

func TestUnitAdd(t *testing.T) {
	u := NewTestUnit()

	n := &Node{ID: big.NewInt(int64(17))}

	normal_n := &Node{ID: big.NewInt(int64(4))}

	if ok := u.add(normal_n); !ok {
		t.Error("Add normal Failed")
	}

	if ok := u.add(n); ok {
		t.Error("Add Failed")
	}
}
func TestUnitGetID(t *testing.T) {
	u := NewTestUnit()
	AddNode(u)

	idx := u.getID(magic7)
	if idx != 3 {
		t.Errorf("Get ID failed with 3 != %d  %v", idx, u.nodes)
	}
	idx = u.getID(big.NewInt(int64(10)))
	if idx != 5 {
		t.Errorf("Get ID failed with 5 != %d  %v", idx, u.nodes[idx])
	}
}

func TestUnitGet(t *testing.T) {

	u := NewTestUnit()
	AddNode(u)

	n := u.Get(magic7)

	if n.ID.Cmp(big.NewInt(int64(10))) == 0 {
		t.Errorf("Get Node failed ")
	}
}

func TestUnitDelete(t *testing.T) {
	u := NewTestUnit()
	AddNode(u)

	if !u.Delete(magic7) {
		t.Error("Delete 9 failed")
	}

	for i, n := range u.nodes {
		if n.ID.Cmp(magic7) == 0 {
			t.Errorf("Delete 9 failed on %d %v", i, u)
		}
	}
}

func TestUnitSuccessorOf(t *testing.T) {
	u := NewTestUnit()
	AddNode(u)

	n := u.successorOf(magic7)
	if n.ID.Cmp(magic7) < 0 {
		t.Errorf("SuccessorOf Failed:%s", n)
	}
	if n.ID.Cmp(big.NewInt(int64(9))) != 0 {
		t.Errorf("Successsor error:%s %s", n, u)
	}
}

func TestUnitUpdateLeader(t *testing.T) {
	u := NewTestUnit()
	AddNode(u)

	if u.Leader == nil || u.Leader.ID.String() != "7" {
		t.Errorf("%v", u)

	}
}
