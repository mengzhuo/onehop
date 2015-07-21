package onehop

import "net/rpc"

type RPCPool struct {
	clients map[string]*rpc.Client
}

func (r *RPCPool) Get(address string) (client *rpc.Client, err error) {
	var ok bool
	if client, ok = r.clients[address]; !ok {
		client, err = rpc.DialHTTP("tcp", address)
	}
	return
}
