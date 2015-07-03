package onehop

import (
	"fmt"
	"testing"
)

func TestNewRoute(t *testing.T) {
	r := NewRoute(4)
	if r.k != 4 {
		t.Error("R.k != 4")
	}
	if r.slices.Len() != 4 {
		t.Errorf("Len Not match :%d", r.slices.Len())
	}
	fmt.Printf("%#v\n", r)

	i := 0
	for elem := r.slices.Front(); elem != nil; elem = elem.Next() {
		slice := elem.Value.(*Slice)
		fmt.Printf("%d  Min %x Max %x\n", i, slice.Min.Bytes(), slice.Max.Bytes())
		i++
	}
}
