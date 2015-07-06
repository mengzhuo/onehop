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

func AddNode(u *Unit) {

	for i := 1; i < 14; i = i + 2 {
		n := &Node{ID: big.NewInt(int64(i))}
		u.Add(n)
	}

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

	if ok := u.Add(normal_n); !ok {
		t.Error("Add normal Failed")
	}

	if ok := u.Add(n); ok {
		t.Error("Add Failed")
	}
}
func TestUnitGetID(t *testing.T) {
	u := NewTestUnit()
	AddNode(u)

	idx := u.GetID(magic7)
	if idx != 3 {
		t.Errorf("Get ID failed with 3 != %d  %v", idx, u.nodes)
	}
	idx = u.GetID(big.NewInt(int64(10)))
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

	n := u.SuccessorOf(magic7)
	if n.ID.Cmp(magic7) < 0 {
		t.Errorf("SuccessorOf Failed:%s", n)
	}
	if n.ID.Cmp(big.NewInt(int64(9))) != 0 {
		t.Errorf("Successsor error:%s %s", n, u)
	}
}

func TestUnitPredecessorOf(t *testing.T) {
	u := NewTestUnit()
	AddNode(u)
	n := u.PredecessorOf(magic7)
	if n == nil {
		t.Fatalf("Can't not get precessor of magic 7")
	}

	if n.ID.Cmp(magic7) >= 0 {
		t.Errorf("PredecessorOf Failed:%s", n)
	}
}
