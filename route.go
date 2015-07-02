// Package onehop provides ...
package onehop

import (
	"container/list"
	"fmt"
)

type Slice struct {
	Leader *Node
	Min    RecordID
	Max    RecordID
	nodes  *list.List
	u      int // the number of units in a slice, u
}

func NewSlice() {

}

type Route struct {
	slices *list.List
	k      int // number of slices the ring is divided into
}

func NewRoute(k int, slice_leader *Node) *Route {

	if k%2 != 0 {
		panic(fmt.Errorf("K should be even number, now are %d", k))
	}

	slices := list.New()
	min := NewFullRecordID()
	min.RShift()
	min.RShift()
	max := NewFullRecordID()

	for i := 0; i < k; i++ {
		s := &Slice{Min: min, Max: max}
		slices.PushBack(s)
		min.RShift()
		max.RShift()
	}

	return &Route{slices, k}

}
