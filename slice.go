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

	i := s.getID(id)

	switch s.Len() - i {
	case 0:
		return nil
	case 1:
		n = s.Nodes[i]
		if n.ID == id {
			return nil
		}
	default:
		n = s.Nodes[i]
		if n.ID == id {
			n = s.Nodes[i+1]
		}
	}
	return
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
	u.RLock()
	defer u.RUnlock()

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

	s.RLock()
	defer s.RUnlock()

	if len(s.Nodes) > 0 {
		return s.Nodes[len(s.Nodes)/2]
	}

	return nil
}

func (s *Slice) Delete(id string) bool {

	i := s.getID(id)

	if i < len(s.Nodes) && s.Nodes[i].ID == id {
		s.Lock()
		s.Nodes = append(s.Nodes[:i], s.Nodes[i+1:]...)
		s.Unlock()
		glog.Infof("slice %s Delete %s", s.Max, id)
		return true
	}
	return false
}

func (s *Slice) Add(n *Node) bool {

	if n.ID < s.Min || n.ID > s.Max {
		return false
	}

	if node := s.Get(n.ID); node == nil {
		glog.Infof("Slice %s add %s", s.Max, n.ID)
		s.Lock()
		s.Nodes = append(s.Nodes, n)
		sort.Sort(s.Nodes)
		s.Unlock()
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
