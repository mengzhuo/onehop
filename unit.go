package onehop

import (
	"math/big"
	"sort"

	"github.com/golang/glog"
)

func NewUnit(min, max *big.Int) *Unit {

	nodes := make(ByID, 0)
	return &Unit{Min: min, Max: max, nodes: nodes}
}

type Unit struct {
	Leader *Node

	Min *big.Int
	Max *big.Int

	nodes ByID
}

func (u *Unit) Get(id *big.Int) (n *Node) {

	i := u.getID(id)
	if i < len(u.nodes) && u.nodes[i].ID.Cmp(id) == 0 {
		// ID in our nodes
		return u.nodes[i]
	}
	return nil
}

func (u *Unit) updateLeader() {
	if u.Len() > 0 {
		u.Leader = u.nodes[u.Len()/2]
		return
	}
	u.Leader = nil
}

func (u *Unit) Delete(id *big.Int) bool {

	i := u.getID(id)
	if i < len(u.nodes) {
		u.nodes = append(u.nodes[:i], u.nodes[i+1:]...)
		u.updateLeader()
		return true
	}
	return false
}

func (u *Unit) add(n *Node) bool {

	if n.ID.Cmp(u.Min) < 0 || n.ID.Cmp(u.Max) > 0 {
		return false
	}

	if i := u.getID(n.ID); i < u.Len() {
		u.nodes[i] = n

	} else {
		u.nodes = append(u.nodes, n)
		sort.Sort(u.nodes)
		u.updateLeader()
	}

	glog.Infof("Unit %x add %x", u.Max, n.ID)
	return true
}

func (u *Unit) getID(id *big.Int) (i int) {

	i = sort.Search(len(u.nodes),
		func(i int) bool {
			return u.nodes[i].ID.Cmp(id) >= 0
		})
	return i
}

func (u *Unit) Len() int {
	return len(u.nodes)
}

func (u *Unit) successorOf(id *big.Int) (n *Node) {

	if u.Len() == 0 {
		// Faster query
		return nil
	}

	i := u.getID(id)
	if i+1 < u.Len() {
		// ID in our nodes
		return u.nodes[i+1]
	}
	return nil
}
