package onehop

import (
	"fmt"
	"net"
	"net/rpc"
	"sync"

	"github.com/golang/glog"
	"github.com/mengzhuo/murmur3"
)

func NewRPCPool() *RPCPool {
	d := murmur3.New128()
	return &RPCPool{make(map[string]*rpc.Client),
		d, &sync.Mutex{}}
}

type RPCPool struct {
	clients map[string]*rpc.Client
	decoder murmur3.Hash128
	*sync.Mutex
}

func (r *RPCPool) Get(address string) (client *rpc.Client, err error) {
	var ok bool
	if client, ok = r.clients[address]; !ok {
		client, err = rpc.Dial("tcp", address)
	}
	return
}

func (r *RPCPool) ToID(key string) string {
	r.Lock()
	defer r.Unlock()
	r.decoder.Reset()

	r.decoder.Write([]byte(key))
	i := r.decoder.Sum(nil)
	return fmt.Sprintf("%x", i)
}

func (s *Service) Get(key string) (i *Item) {
	return s.GetID(s.RPCPool.ToID(key))
}

func (s *Service) GetID(key string) *Item {

	items := make([]*Item, 0, s.R)
	var first_node *Node
	id := key

	for i := 0; i < s.R; i++ {

		node := s.route.SuccessorOf(id)
		glog.V(3).Infof("Get Found SuccessorOf %s is %s", id, node)

		if node == nil {
			// no node in this slice
			continue
		}

		// Go to next node.ID
		id = node.ID

		if first_node == nil {
			first_node = node
		} else if node == first_node {
			// Loopbacked
			break
		}

		if node.ID == s.id {
			// it's ourself
			s.DB.mu.RLock()

			if item, ok := s.DB.db[key]; ok {
				items = append(items, item)
			}
			s.DB.mu.RUnlock()
			continue
		}

		client, err := s.RPCPool.Get(node.Addr.String())
		if s.RPCError(err, node) {
			continue
		}
		var reply *Item
		err = client.Call("Storage.Get", key, &reply)
		if s.RPCError(err, node) {
			continue
		}
		glog.V(1).Infof("Get %s From:%s", key, node.Addr)
		if err == nil && reply != nil {
			glog.V(1).Infof("Get reply %s", reply)
			items = append(items, reply)
		}
	}
	var max_item *Item
	if len(items) != 0 {
		max_item = items[0]
		for _, item := range items[1:] {
			if item.Ver > max_item.Ver {
				max_item = item
			}
		}
	}
	return max_item
}

func (s *Service) Put(key string, item *Item) (count int) {

	return s.PutID(s.RPCPool.ToID(key), item)
}

func (s *Service) PutID(key string, item *Item) (count int) {

	var first_node *Node
	id := key
	count = 0

	for i := 0; i < s.W; i++ {

		node := s.route.SuccessorOf(id)
		glog.V(3).Infof("Put Found SuccessorOf %s is %s", id, node)

		if node == nil {
			// no more to write...
			continue
		}

		// Go to next node.ID
		id = node.ID

		if first_node == nil {
			first_node = node
		} else if node == first_node {
			// Loopbacked
			break
		}

		if node.ID == s.id {
			// it's ourself
			if s.DB.put(key, item) {
				count += 1
			}
			continue
		}
		glog.V(1).Infof("Put %s to %s", key, node.Addr.String())

		client, err := s.RPCPool.Get(node.Addr.String())
		if s.RPCError(err, node) {
			continue
		}

		args := &PutArgs{key, item}
		var reply *bool
		err = client.Call("Storage.Put", args, &reply)
		if !s.RPCError(err, node) {
			count += 1
		}
	}
	return
}

func (s *Service) RPCError(err error, node *Node) bool {
	if err != nil {
		switch err.(type) {
		case *net.OpError:
			glog.Errorf("RPC call failed delete Node:%s", node.ID)
			s.AddExchangeEvent(&Event{node.ID, LEAVE, node.Addr})
			s.route.Delete(node.ID)
		}
		return true
	}
	return false
}
