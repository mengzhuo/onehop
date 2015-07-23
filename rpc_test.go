package onehop

import (
	"net/rpc"
	"testing"
)

func TestStoreGet(t *testing.T) {

	r := RPCPool{make(map[string]*rpc.Client, 0)}
	client, err := r.Get("10.5.4.152:7676")
	if err != nil {
		t.Error(err)
	}
	var reply *Item
	err = client.Call("Storage.Get", []byte{1, 2, 3, 4}, reply)
	if err != nil {
		t.Error(err)
	}
}
