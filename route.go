// Package onehop provides ...
package onehop

import (
	"container/list"
)

type Slice struct {
	Leader *Node
	Min    RecordID
	Max    RecordID
	nodes  *list.List
	u      int // the number of units in a slice, u
}

type Route struct {
	slices *list.List
	k      int // number of slices the ring is divided into
}

func NewRoute(k int) *Route {
	return &Route{}
}
