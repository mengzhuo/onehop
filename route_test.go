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
	if r.div != 2 {
		t.Error(r.div)
	}

	r = NewRoute(257)
	if r.div != 4 {
		t.Error(r.div)
	}
}

func TestGetIndex(t *testing.T) {

	r := NewRoute(257)
	id := "f2ffffffffffffffffffffffffffffff"
	idx := r.GetIndex(id)
	if s := r.slices[idx]; s.Min > id || s.Max < id {
		t.Error(idx, s)
	}
	id = FULL_ID
	idx = r.GetIndex(id)
	if s := r.slices[idx]; s.Min > id || s.Max < id {
		t.Error(idx, s)
	}

	r = NewRoute(2)
	i := r.GetIndex("2bf2b904cf874746c1a0f8626af0412b")
	if i != 0 {
		t.Error(r.slices, i)
	}
}

func BenchmarkGetIndex(b *testing.B) {
	r := NewRoute(257)
	for i := 0; i < b.N; i++ {
		r.GetIndex(FULL_ID)
	}

}

func TestRouteAdd(t *testing.T) {

	r := NewRoute(4)
	for i := 0; i < 256; i += 15 {
		id := fmt.Sprintf("%02x", []byte{byte(i)}) + FULL_ID[2:]
		r.Add(&Node{id, nil, time.Now().Unix(), nil})
	}

	if r.Len() != 18 {
		for _, slice := range r.slices {
			t.Errorf("Slice %s->%s :%d\n", slice.Min, slice.Max, slice.Len())
			for _, n := range slice.Nodes {
				t.Error("|- " + n.String())
			}
		}
	}

	r = NewRoute(17)
	for i := 0; i < 256; i += 2 {
		id := fmt.Sprintf("%02x", []byte{byte(i)}) + FULL_ID[2:]
		r.Add(&Node{id, nil, time.Now().Unix(), nil})
	}
	if r.Len() != 128 {
		for _, slice := range r.slices {
			t.Errorf("Slice %s->%s :%d\n", slice.Min, slice.Max, slice.Len())
			for _, n := range slice.Nodes {
				t.Error("|- " + n.String())
			}
		}
	}
}

func TestRouteDelete(t *testing.T) {

	r := NewRoute(4)
	for i := 0; i < 256; i += 15 {
		id := fmt.Sprintf("%02x", []byte{byte(i)}) + FULL_ID[2:]
		r.Add(&Node{id, nil, time.Now().Unix(), nil})
	}
	r.Delete("f0ffffffffffffffffffffffffffffff")
	if r.Len() != 17 {
		for _, slice := range r.slices {
			t.Errorf("Slice %s->%s :%d\n", slice.Min, slice.Max, slice.Len())
			for _, n := range slice.Nodes {
				t.Error("|- " + n.String())
			}
		}
	}
}

func TestRouteSuccessorOf(t *testing.T) {
	/*
		r := NewRoute(4)
		for i := 0; i < 256; i += 15 {
			id := fmt.Sprintf("%02x", []byte{byte(i)}) + FULL_ID[2:]
			r.Add(&Node{id, nil, time.Now().Unix(), nil})
		}
		n := r.SuccessorOf("f0ffffffffffffffffffffffffffffff")
		if n == nil {
			t.Error(n, "not found")
		}
		n = r.SuccessorOf(FULL_ID)
		if n == nil {
			t.Error(n, "not found")
		}
	*/
	r := NewRoute(7)
	for i := 0; i < 20; i += 15 {
		id := fmt.Sprintf("%02x", []byte{byte(i)}) + FULL_ID[2:]
		r.Add(&Node{id, nil, time.Now().Unix(), nil})
	}
	n := r.SuccessorOf("fffffffffffffffffffffffffffffffe")
	if n == nil || n.ID != "00ffffffffffffffffffffffffffffff" {
		t.Error("expecting 00ffffffffffffffffffffffffffffff\nbut found...........", n.ID)
		for _, n := range r.slices[0].Nodes {
			t.Log(n.ID)
		}
	}
}
