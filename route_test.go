package onehop

import (
	"fmt"
	"testing"
)

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
