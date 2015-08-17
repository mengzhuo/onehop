package onehop

import (
	"fmt"
	"testing"
)

func TestLeakyBuffer(t *testing.T) {

	buf := NewLeakyBuffer(1024, 1024)
	p := buf.Get()
	p[2] = 0x8
	pn := buf.Get()

	if fmt.Sprintf("%#p", p) == fmt.Sprintf("%#p", pn) {
		t.Error("Get same byte")
	}

	buf.Put(p)
	pnn := buf.Get()

	if fmt.Sprintf("%#p", p) != fmt.Sprintf("%#p", pnn) {
		t.Error("Get difference byte")
	}

}

func BenchmarkLeakyBufferGet(b *testing.B) {
	b.StopTimer()
	buf := NewLeakyBuffer(1024, 1024)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p := buf.Get()
			buf.Put(p)
		}
	})
}
