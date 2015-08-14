package onehop

import (
	"fmt"
	"testing"
	"time"
)

func TestNewRoute(t *testing.T) {

	r := NewRoute(16)
	if len(r.slices) != 16 {
		t.Error(r.slices)
	}
	if r.div != 1 {
		t.Error(r.div)
	}
	r = NewRoute(4)
	if len(r.slices) != 4 {
		t.Error(r.slices)
	}
	if r.div != 1 {
		t.Error(r.div)
	}
	for _, s := range r.slices {
		fmt.Println(s.Min, s.Max)
	}
}

func TestRouteAdd(t *testing.T) {

	r := NewRoute(4)
	for id := 0; id < 32; id += 2 {

		r.Add(&Node{BytesToId([]byte{byte(id)}),
			nil, time.Now()})
	}
	r.Add(&Node{FULL_ID, nil, time.Now()})
	for _, n := range r.slices[0].Nodes {
		fmt.Println(n.String())
	}

	if r.Len() != 17 {
		t.Error(r.Len())
	}
}
