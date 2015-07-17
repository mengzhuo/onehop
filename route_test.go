package onehop

import (
	"fmt"
	"math/big"
	"testing"
)

func AddRouteNode(r *Route) {

	for i := 14; i > 0; i = i - 2 {
		n := &Node{ID: big.NewInt(int64(i))}
		r.Add(n)
	}
}

func TestNewRoute(t *testing.T) {
	r := NewRoute(4)
	AddRouteNode(r)
	if r.k != 4 {
		t.Error("R.k != 4")
	}
	if len(r.slices) != 4 {
		t.Errorf("Len Not match :%d", len(r.slices))
	}
	fmt.Printf("%#v\n", r)

	if len(r.slices[0].nodes) != 7 {
		t.Errorf("Len Not match :%d", len(r.slices[0].nodes))
	}

	for i, slice := range r.slices {
		fmt.Printf("%d Leader:%s  Min %0x Max %0x\n", i, slice.Leader, slice.Min.Bytes(), slice.Max.Bytes())
		for j := 0; j < len(slice.nodes); j++ {
			u := slice.nodes[j]
			fmt.Printf("|---  %d ID %016X \n", j, u.ID)
		}
		i++
	}

	if r.slices[0].Leader == nil || r.slices[0].Leader.ID.Uint64() != uint64(8) {
		t.Errorf("No Leader :%d", len(r.slices[0].nodes))
	}
}
