package onehop

import "testing"

func TestNewMsg(t *testing.T) {

	header := []byte{0x2, 0x3, 0x4, 0x5}
	h := ParseHeader(header)
	if h.ID != 2 || h.Type != KEEP_ALIVE || h.Count != 1029 {
		t.Error(h)
	}
}

func BenchmarkNewMsg(b *testing.B) {

	header := []byte{0x2, 0x3, 0x4, 0x5}

	for i := 0; i < b.N; i++ {
		_ = ParseHeader(header)
	}
}
