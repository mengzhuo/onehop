package onehop

import (
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/golang/glog"
)

type Service struct {
	conn *net.UDPConn

	route *Route

	id string

	exchangeEvent []Event
	notifyEvent   []Event

	selfSlice   *Slice
	pinger      *Node
	leftPonger  *Node
	rightPonger *Node
	W, R        int
	//RPCPool     *RPCPool
}

// NetType, Address for UDP connection
// k for OneHop slice number
func NewService(netType, address string, k, w, r int) *Service {
	glog.Infof("Write:%d, Read:%d", k, r)
	listener, err := newListener(netType, address)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	glog.Infof("Listening to :%s %s", netType, address)
	route = NewRoute(k)
	max := new(big.Int).SetBytes(FullID)

	vid, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}

	id := BytesToId(vid.Bytes())
	glog.Infof("initial id:%s", id)

	n := &Node{ID: id,
		Addr: listener.LocalAddr().(*net.UDPAddr)}
	route.Add(n)

	slice := route.slices[route.GetIndex(id)]

	service = &Service{
		listener, route,
		id,
		make([]Event, 0), make([]Event, 0), slice, n,
		nil, nil, w, r}

	gob.Register(Msg{})
	gob.Register(Event{})
	glog.V(3).Infof("RPC Listener Accepted")
	return service
}

/*
func (s *Service) Get(key []byte) *Item {

	items := make([]*Item, 0)
	var first_node *Node
	id := BytesToId(key)

	for i := 0; i < s.R; i++ {
		node := s.route.SuccessorOf(id)

		glog.V(3).Infof("Get Found SuccessorOf %x is %s", id, node)
		if node == nil {
			// no more to read...
			continue
		}
		if first_node == nil {
			first_node = node
		} else {
			// Loopbacked
			if node == first_node {
				break
			}
		}

		if node == s.selfNode {
			// it's ourself
			if item, ok := s.db.db[string(key)]; ok {
				items = append(items, item)
			}
			id = node.ID
			continue
		}

		client, err := s.RPCPool.Get(node.Addr.String())

		if err != nil {
			glog.Error(err)
			s.NotifySliceLeader(node, LEAVE)
			s.route.Delete(node.ID)
			continue
		}
		var reply *Item
		err = client.Call("Storage.Get", key, &reply)
		glog.V(1).Infof("Get %x From:%s", key, node.Addr)
		if err == nil && reply != nil {
			glog.V(1).Infof("Get reply %s", reply)
			items = append(items, reply)
		}
		id = node.ID
	}

	if len(items) == 0 {
		return nil
	}

	max_item := items[0]

	if len(items) > 1 {
		for _, item := range items[1:] {
			if item.Ver > max_item.Ver {
				max_item = item
			}
		}
	}

	return max_item
}

func (s *Service) PutByString(key string, item *Item) int {
	k, err := hex.DecodeString(key)
	if err != nil {
		glog.Error(err)
		return 0
	}
	return s.Put(k, item)
}

func (s *Service) Put(key []byte, item *Item) (count int) {

	id := BytesToId(key)

	var first_node *Node
	count = 0

	for i := 0; i < s.W; i++ {

		node := s.route.SuccessorOf(id)
		glog.V(3).Infof("Put Found SuccessorOf %x is %s", id, node)

		if node == nil {
			// no more to write...
			continue
		}

		if first_node == nil {
			first_node = node
		} else {
			// Loopbacked
			if node == first_node {
				break
			}
		}
		if node == s.selfNode {
			// it's ourself
			k := string(key)
			if selfItem, ok := s.db.db[k]; !ok {
				s.db.db[k] = item
				count += 1
			} else {
				if selfItem.Ver < item.Ver {
					s.db.db[k] = item
					count += 1
				}
			}
			id = node.ID
			continue
		}
		glog.V(1).Infof("Put %x to %s", key, node.Addr.String())
		id = node.ID

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
*/
func (s *Service) RPCError(err error, node *Node) bool {
	if err != nil {
		glog.Errorf("Node:%x %#v", node.ID, err)
		switch err.(type) {

		case *net.OpError:
			glog.Error("RPC call failed")
			s.NotifySliceLeader(node, LEAVE)
			s.route.Delete(node.ID)
		}
		return true
	}
	return false
}

func (s *Service) BootStrapFrom(address string) {

	glog.Infof("BootStrap From :%s", address)
	addr, _ := net.ResolveUDPAddr("udp", address)
	msg := new(Msg)
	msg.Type = BOOTSTRAP
	msg.NewID()
	s.SendMsg(addr, msg)
}

func BytesToId(p []byte) string {
	return fmt.Sprintf("%032x", p)
}

func (s *Service) Listen() {

	for {

		//p := s.bytePool.Get()
		p := make([]byte, 1024)
		n, addr, err := s.conn.ReadFromUDP(p)
		if err != nil || n < 3 {
			glog.Errorf("insufficient data for parsing...", addr, p[:n])
			continue
		}

		msg := new(Msg)
		err = json.Unmarshal(p[:n], msg)
		if err != nil {
			panic(err)
		}
		glog.V(10).Infof("Recv From:%x TYPE:%s", msg.From, typeName[msg.Type])
		//s.Handle(addr, msg)
	}
}

func (s *Service) Send(dstAddr *net.UDPAddr, p []byte) {

	defer func() {
		if r := recover(); r != nil {
			glog.ErrorDepth(0, r)
		}
	}()
	if dstAddr == s.conn.LocalAddr() {
		return
	}
	s.conn.WriteToUDP(p, dstAddr)
}

func (s *Service) SendMsg(dstAddr *net.UDPAddr, msg *Msg) {

	p, err := json.Marshal(msg)
	if err != nil {
		glog.Errorf("Error on parse %s", msg)
		return
	}
	s.Send(dstAddr, p)
}

func (s *Service) NotifySliceLeader(n *Node, status byte) {

	slice := s.selfSlice

	if slice.Leader() == nil {
		glog.Fatal("Something is wrong!!! we are still in the slice and there is no slice leader?")
	}
	glog.V(5).Infof("Notify Slice leader about  %s", n)
	e := Event{n.ID, time.Now(), status, n.Addr.String()}

	s.exchangeEvent = append(s.exchangeEvent, e)

}
