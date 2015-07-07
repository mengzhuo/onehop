package onehop

import (
	"fmt"
	"math/big"
	"testing"
)

func AddNode(u *Unit) {

	for i := 1; i < 14; i = i + 2 {
		n := &Node{ID: big.NewInt(int64(i))}
		u.Add(n)
	}

}

func TestNewRoute(t *testing.T) {
	r := NewRoute(4, 4)
	if r.k != 4 {
		t.Error("R.k != 4")
	}
	if len(r.slices) != 4 {
		t.Errorf("Len Not match :%d", len(r.slices))
	}
	fmt.Printf("%#v\n", r)

	for i, slice := range r.slices {
		fmt.Printf("%d  Min %0x Max %0x\n", i, slice.Min.Bytes(), slice.Max.Bytes())
		for j := 0; j < len(slice.units); j++ {
			u := slice.units[j]
			fmt.Printf("|---  %d, Min %016X Max %016X\n", j, u.Min, u.Max)
		}
		i++
	}
}

func TestRouteGetSliceIndex(t *testing.T) {
	r := NewRoute(8, 10)

	sidx, uidx := r.GetIndex(new(big.Int).SetBytes(fullByte))
	if 7 != sidx || 9 != uidx {
		t.Errorf("GetSliceIdx failed slice index %d, unit index", sidx, uidx)
	}
	inter := make([]byte, 0)
	inter = append(inter, fullByte...)
	inter[0] = 0xc0

	sidx, uidx = r.GetIndex(new(big.Int).SetBytes(inter))
	if 6 != sidx || uidx != 0 {
		t.Errorf("GetSliceIdx failed slice index %d, unit index", sidx, uidx)
	}
}

func TestRouteSuccessorOf(t *testing.T) {
	r := NewRoute(8, 10)
}
