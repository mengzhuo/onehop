package onehop

import (
	"sort"
	"sync"

	"github.com/golang/glog"
)

type Slice struct {
	Min   string
	Max   string
	Nodes ByNodeID
	*sync.Mutex
}

func (s *Slice) Len() int {
	return len(s.nodes)
}

func (s *Slice) successorOf(id string) (n *Node) {

	if s.Len() == 0 {
		// Faster query
		return nil
	}

	i := s.getID(id)

	if i >= s.Len()-1 {
		return nil
	}
	return s.nodes[i+1]
}

func (s *Slice) predecessorOf(id string) (n *Node) {

	if s.Len() == 0 {
		// Faster query
		return nil
	}

	i := s.getID(id)

	if i != 0 {
		return s.Nodes[i-1]
	}
	return nil
}

func (u *Slice) getID(id string) (i int) {

	i = sort.Search(len(u.nodes),
		func(i int) bool {
			return u.nodes[i] >= id
		})
	return i
}
func (u *Slice) Get(id string) (n *Node) {

	i := u.getID(id)
	if i < len(u.nodes) && u.nodes[i].ID == id {
		// ID in our nodes
		return u.nodes[i]
	}
	return nil
}

func (u *Slice) Leader() *Node {

	if len(u.nodes) > 0 {
		return u.nodes[len(u.nodes)/2]
	}

	return nil
}

func (s *Slice) Delete(id string) bool {

	s.Lock()
	defer s.Unlock()

	i := u.getID(id)

	if i < len(s.nodes) && s.nodes[i].ID == id {
		s.nodes = append(s.nodes[:i], s.nodes[i+1:]...)
		glog.Infof("slice %s Delete %s", s.Max, id)
		return true
	}
	return false
}

func (s *Slice) Add(n *Node) bool {

	s.Lock()
	defer s.Unlock()

	if n.ID < s.Min || n.ID > s.Max {
		return false
	}

	if node := s.Get(n.ID); node == nil {
		glog.Infof("Slice %s add %s", s.Max, n.ID)
		s.nodes = append(s.nodes, n)
		sort.Sort(s.nodes)
	}
	return true
}

type ByNodeID []*Node

func (b ByNodeID) Len() int {
	return len(b)
}

func (b ByNodeID) Less(i, j int) bool {
	return b[i] < b[j]
}

func (b ByNodeID) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
