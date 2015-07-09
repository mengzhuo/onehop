package onehop

import "testing"

func TestKeepAliceMsgunpack(t *testing.T) {

	h := &MessageHeader{0x1, KEEP_ALIVE, uint32(4)}

	if h.ID != uint32(4) {
		t.Error("header error")
	}
}
