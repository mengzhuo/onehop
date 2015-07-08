package onehop

import (
	"fmt"
	"math/big"
	"testing"
)

func AddNode(u *Unit) {

	for i := 1; i < 14; i = i + 2 {
		n := &Node{ID: big.NewInt(int64(i))}
		u.add(n)
	}

}

func AddRouteNode(r *Route) {

	for i := 1; i < 14; i = i + 2 {
		n := &Node{ID: big.NewInt(int64(i))}
		r.Add(n)
	}
}

func TestNewRoute(t *testing.T) {
	r := NewRoute(4, 4)
	AddRouteNode(r)
	if r.k != 4 {
		t.Error("R.k != 4")
	}
	if len(r.slices) != 4 {
		t.Errorf("Len Not match :%d", len(r.slices))
	}
	fmt.Printf("%#v\n", r)

	for i, slice := range r.slices {
		fmt.Printf("%d Leader:%s  Min %0x Max %0x\n", i, slice.Leader, slice.Min.Bytes(), slice.Max.Bytes())
		for j := 0; j < len(slice.units); j++ {
			u := slice.units[j]
			fmt.Printf("|---  %d, Leader:%s Min %016X Max %016X\n", j, u.Leader, u.Min, u.Max)
		}
		i++
	}
}

func TestRouteGetSliceIndex(t *testing.T) {
	r := NewRoute(8, 10)

	sidx, uidx := r.GetIndex(new(big.Int).SetBytes(FullID))
	if 7 != sidx || 9 != uidx {
		t.Errorf("GetSliceIdx failed slice index %d, unit index", sidx, uidx)
	}
	inter := make([]byte, 0)
	inter = append(inter, FullID...)
	inter[0] = 0xc0

	sidx, uidx = r.GetIndex(new(big.Int).SetBytes(inter))
	if 6 != sidx || uidx != 0 {
		t.Errorf("GetSliceIdx failed slice index %d, unit index", sidx, uidx)
	}
}

func TestRouteForward(t *testing.T) {
	r := NewRoute(8, 10)
	AddNode(r.slices[0].units[0])

	n := r.SuccessorOf(big.NewInt(int64(99)))
	if n == nil {
		t.Errorf("Route SuccessorOf failed %s", n)
	}
}

func TestRouteUpdateLeader(t *testing.T) {
	r := NewRoute(8, 10)
	AddRouteNode(r)

	n := r.slices[0].Leader
	if n == nil || n.ID.Cmp(big.NewInt(int64(7))) != 0 {
		t.Errorf("Route Leader failed %s", n)
	}
}
