// Package onehop provides ...
package onehop

import (
	"container/list"
)

type Route struct {
	ids   map[RecordID]*Node
	table list.List

	k int // number of slices the ring is divided into
	u int // the number of units in a slice
}

func (r *Route) LookUpByString(id string) (n *Node, existed bool) {

	rid := NewIDFromString(id)
	return r.Get(rid)
}

func (r *Route) LookUp(id RecordID) (n *Node, existed bool) {

	t, _ := r.FetchTree(id)
	for _, n := range t.NodeList {
		if n.ID.Equal(id) {
			return n, true
		}

	}
	return nil, false
}

func (r *Route) Get(id RecordID) (n *Node, existed bool) {

	n, existed = r.ids[id]
	return
}

func (r *Route) Put(n *Node) {

	if _, existed := r.Get(n.ID); existed {
		return
	}

	if t, elem := r.FetchTree(n.ID); t != nil {

		if len(t.NodeList) > t.Limit {
			left := t.SplitBy(n)
			r.table.InsertBefore(left, elem)
			return
		}
		t.NodeList = append(t.NodeList, n)
	}
}

func (r *Route) FetchTree(id RecordID) (t *Tree, elem *list.Element) {

	for elem := r.table.Front(); elem != nil; elem = elem.Next() {
		t := elem.Value.(*Tree)
		if t.Min.Less(id) && id.Less(t.Max) {
			return t, elem
		}
	}
	return nil, nil
}

func (r *Route) SuccessorOf(id RecordID) (n *Node, existed bool) {

	if t, elem := r.FetchTree(id); t != nil {
		for _, n = range t.NodeList {
			if id.Less(n.ID) {
				existed = true
				return
			}
		}

		for elem = elem.Next(); elem != nil; elem = elem.Next() {
			for _, n = range t.NodeList {
				if id.Less(n.ID) {
					existed = true
					return
				}
			}

		}
	}
	return
}

func (r *Route) PredecessorOf(id RecordID) (n *Node, existed bool) {

	if t, elem := r.FetchTree(id); t != nil {
		for _, n = range t.NodeList {
			if n.ID.Less(id) {
				existed = true
				return
			}
		}

		for elem = elem.Prev(); elem != nil; elem = elem.Prev() {
			for _, n = range t.NodeList {
				if n.ID.Less(id) {
					existed = true
					return
				}
			}

		}
	}
	return

}
