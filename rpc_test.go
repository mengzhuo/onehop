package onehop

/*
func TestStorePut(t *testing.T) {

	r := RPCPool{make(map[string]*rpc.Client, 0)}

	client, err := r.Get("10.5.4.152:7676")
	if err != nil {
		t.Error(err)
	}

	item := &Item{6, []byte("ok")}
	args := &PutArgs{[]byte("ok"), item}
	reply := new(bool)
	err = client.Call("Storage.Put", args, reply)

	if err != nil {
		t.Error(err)
	}
}

func TestStoreGet(t *testing.T) {

	r := RPCPool{make(map[string]*rpc.Client, 0)}

	client, err := r.Get("10.5.4.152:7676")
	if err != nil {
		t.Error(err)
	}

	key := []byte("id")

	item := &Item{6, []byte("Data")}
	args := &PutArgs{key, item}

	reply := new(bool)
	err = client.Call("Storage.Put", args, reply)

	if err != nil || !*reply {
		t.Error(err)
	}

	get := new(Item)
	err = client.Call("Storage.Get", key, &get)
	if err != nil {
		t.Error(err)

	}

	if get.Ver != 6 || (string(get.Data) != string([]byte("Data"))) {
		t.Errorf("Error on get %x %s", key, get)
	}
}

/*
func TestStoreDelete(t *testing.T) {
	r := RPCPool{make(map[string]*rpc.Client, 0)}

	client, err := r.Get("10.5.4.152:7676")
	if err != nil {
		t.Error(err)
	}

	key := []byte("delete test")
	item := &Item{6, []byte("Data")}
	args := &PutArgs{key, item}
	reply := new(bool)
	err = client.Call("Storage.Put", args, reply)

	reply = new(bool)
	dargs := &DeleteArgs{key, item.Ver}
	err = client.Call("Storage.Delete", dargs, reply)
	if err != nil {
		t.Error(err)

	}

}
*/
