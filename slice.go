package onehop

import (
	"sort"
	"sync"

	"github.com/golang/glog"
)

func NewSlice(min, max string) *Slice {
	nodes := make(ByNodeID, 0)
	return &Slice{min, max, nodes, &sync.RWMutex{}}
}

type Slice struct {
	Min   string
	Max   string
	Nodes ByNodeID
	*sync.RWMutex
}

func (s *Slice) Len() int {
	return len(s.Nodes)
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
	return s.Nodes[i+1]
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

	i = sort.Search(len(u.Nodes),
		func(i int) bool {
			return u.Nodes[i].ID >= id
		})
	return i
}
func (s *Slice) Get(id string) (n *Node) {

	i := s.getID(id)
	if i < len(s.Nodes) && s.Nodes[i].ID == id {
		// ID in our nodes
		return s.Nodes[i]
	}
	return nil
}

func (s *Slice) Leader() *Node {

	if len(s.Nodes) > 0 {
		return s.Nodes[len(s.Nodes)/2]
	}

	return nil
}

func (s *Slice) Delete(id string) bool {

	s.Lock()
	defer s.Unlock()

	i := s.getID(id)

	if i < len(s.Nodes) && s.Nodes[i].ID == id {
		s.Nodes = append(s.Nodes[:i], s.Nodes[i+1:]...)
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
		s.Nodes = append(s.Nodes, n)
		sort.Sort(s.Nodes)
	}
	return true
}

type ByNodeID []*Node

func (b ByNodeID) Len() int {
	return len(b)
}

func (b ByNodeID) Less(i, j int) bool {
	return b[i].ID < b[j].ID
}

func (b ByNodeID) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
