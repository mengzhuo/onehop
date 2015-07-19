package onehop

import (
	"math/big"
	"sort"

	"github.com/golang/glog"
)

type Slice struct {
	Leader *Node
	Min    *big.Int
	Max    *big.Int
	nodes  ByID
}

func (s *Slice) Len() int {
	return len(s.nodes)
}

func (s *Slice) successorOf(id *big.Int) (n *Node) {

	if s.Len() == 0 {
		// Faster query
		return nil
	}

	i := s.getID(id)
	if i+1 < s.Len() {
		// ID in our nodes
		return s.nodes[i+1]
	}
	return nil
}

func (s *Slice) predecessorOf(id *big.Int) (n *Node) {

	if s.Len() == 0 {
		// Faster query
		return nil
	}

	i := s.getID(id)
	if i > 0 {
		// ID in our nodes
		return s.nodes[i-1]
	}
	return nil
}
func (u *Slice) getID(id *big.Int) (i int) {

	i = sort.Search(len(u.nodes),
		func(i int) bool {
			return u.nodes[i].ID.Cmp(id) >= 0
		})
	return i
}
func (u *Slice) Get(id *big.Int) (n *Node) {

	i := u.getID(id)
	if i < len(u.nodes) && u.nodes[i].ID.Cmp(id) == 0 {
		// ID in our nodes
		return u.nodes[i]
	}
	return nil
}

func (u *Slice) updateLeader() {
	if u.Len() > 0 {
		u.Leader = u.nodes[u.Len()/2]
		return
	}
	u.Leader = nil
}

func (u *Slice) Delete(id *big.Int) bool {

	i := u.getID(id)

	if i < len(u.nodes) && u.nodes[i].ID.Cmp(id) == 0 {
		u.nodes = append(u.nodes[:i], u.nodes[i+1:]...)
		u.updateLeader()
		return true
	}
	glog.Infof("slice %x Delete %x", u.Max, id)
	return false
}

func (u *Slice) add(n *Node) bool {

	if n.ID.Cmp(u.Min) < 0 || n.ID.Cmp(u.Max) > 0 {
		return false
	}
	if node := u.Get(n.ID); node == nil {
		glog.Infof("Unit %x add %x", u.Max, n.ID)
		u.nodes = append(u.nodes, n)
		sort.Sort(u.nodes)
	}
	u.updateLeader()

	return true
}

type ByID []*Node

func (b ByID) Len() int           { return len(b) }
func (b ByID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByID) Less(i, j int) bool { return b[i].ID.Cmp(b[j].ID) < 0 }
