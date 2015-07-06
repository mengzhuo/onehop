package onehop

import (
	_ "fmt"
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

func NewUnit(min, max *big.Int) *Unit {

	nodes := make(ByID, 0)
	return &Unit{Min: min, Max: max, nodes: nodes}
}

type Unit struct {
	Min *big.Int
	Max *big.Int

	nodes ByID
}

func (u *Unit) Add(n *Node) bool {

	if n.ID.Cmp(u.Min) < 0 || n.ID.Cmp(u.Max) > 0 {
		return false
	}

	u.nodes = append(u.nodes, n)
	sort.Sort(u.nodes)
	return true
}

func (u *Unit) GetID(id *big.Int) (i int) {

	i = sort.Search(len(u.nodes),
		func(i int) bool {
			return u.nodes[i].ID.Cmp(id) >= 0
		})
	return i
}

func (u *Unit) Get(id *big.Int) (n *Node) {

	i := u.GetID(id)
	if i < len(u.nodes) && u.nodes[i].ID.Cmp(id) == 0 {
		// ID in our nodes
		return u.nodes[i]
	}
	return nil
}

func (u *Unit) Delete(id *big.Int) bool {

	i := u.GetID(id)
	if i < len(u.nodes) {
		u.nodes = append(u.nodes[:i], u.nodes[i+1:]...)
		return true
	}
	return false
}

func (u *Unit) Len() int {
	return len(u.nodes)
}

func (u *Unit) SuccessorOf(id *big.Int) (n *Node) {

	i := sort.Search(len(u.nodes), func(i int) bool { return u.nodes[i].ID.Cmp(id) >= 0 })
	if i < len(u.nodes) {
		// ID in our nodes
		return u.nodes[i]
	}
	return nil
}

func (u *Unit) PredecessorOf(id *big.Int) (n *Node) {

	i := sort.Search(len(u.nodes),
		func(i int) bool {
			return id.Cmp(u.nodes[i].ID) >= 0
		})
	if i < len(u.nodes) {
		// ID in our nodes
		return u.nodes[i]
	}
	return nil
}
