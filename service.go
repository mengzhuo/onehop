package onehop

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/golang/glog"
)

var ALIVE_EVENT *Event

type Service struct {
	conn *net.UDPConn

	route *Route

	id string

	exchangeEvent []*Event
	exchangeLock  *sync.RWMutex
	notifyEvent   []*Event
	notifyLock    *sync.RWMutex

	selfSlice *Slice
	W, R      int
	bytePool  *LeakyBuffer
}

func NewUDPDecoder(buf *bytes.Buffer) *json.Decoder {
	return json.NewDecoder(buf)
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
	vid := make([]byte, 16)
	rand.Read(vid)
	id := hex.EncodeToString(vid)

	n := &Node{id,
		listener.LocalAddr().(*net.UDPAddr),
		time.Now().Unix(),
		&sync.Mutex{}}
	route.Add(n)

	slice := route.slices[route.GetIndex(id)]
	glog.V(1).Info("Self Slice:", slice.Max)

	service = &Service{
		listener, route,
		id,
		make([]*Event, 0),
		&sync.RWMutex{},
		make([]*Event, 0),
		&sync.RWMutex{},
		slice,
		w, r,
		NewLeakyBuffer(1024, MSG_MAX_SIZE),
	}

	glog.V(3).Infof("RPC Listener Accepted")
	go service.Start()
	ALIVE_EVENT = &Event{id, JOIN, listener.LocalAddr().(*net.UDPAddr)}
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
			//s.NotifySliceLeader(node, LEAVE)
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
	msg.From = s.id
	s.SendMsg(addr, msg)
}

func BytesToId(p []byte) string {
	return fmt.Sprintf("%032x", p)
}

func (s *Service) Listen() {

	for {

		p := s.bytePool.Get()
		n, addr, err := s.conn.ReadFromUDP(p)
		if err != nil || n < 5 {
			glog.Errorf("insufficient data from %s:%x", addr, p[:n])
			continue
		}
		glog.V(10).Infof("Recv From %s with %d", addr, n)

		msg := new(Msg)

		err = json.Unmarshal(p[:n], msg)
		if err != nil {
			glog.Error(err, p[:n])
			continue
		}

		glog.V(9).Infof("Msg:%s", msg)
		s.handle(addr, msg)
		s.bytePool.Put(p)
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
		glog.Error(err)
		return
	}
	glog.V(10).Infof("SEND %s with %x", dstAddr, p)
	s.Send(dstAddr, p)
}

func (s *Service) Start() {

	glog.Info("Checker on")
	for {
		time.Sleep(1 * time.Second)
		s.check()
	}
}

func (s *Service) Status() {
	glog.Infof("-------------NODE:%s------------", s.id)
	for _, slice := range s.route.slices {
		slice.RLock()
		glog.Infof("Slice[%d]:%s->%s", slice.Len(), slice.Min, slice.Max)
		l := slice.Leader()
		for i, n := range slice.Nodes {
			n.Lock()
			p := "├━ %s %s"
			leader := ""
			if i == slice.Len()-1 {
				p = "└━ %s %s"
			}
			if n.ID == l.ID {
				leader = "[LEADER]"
			}
			glog.Infof(p, n.ID, leader)
			n.Unlock()
		}
		slice.RUnlock()
	}
}

func (s *Service) checkSelfSlice(slice *Slice, now int64) {

	timeouted := make([]string, 0)
	slice.RLock()
	leader := slice.Leader()

	for _, n := range slice.Nodes {

		n.Lock()
		switch n.ID {
		case s.id:
		//pass
		case leader.ID:
			if now-n.updateAt > SLICE_LEADER_TIMEOUT {
				glog.Infof("Leader Node:%s timeouted", n.ID)
				s.AddExchangeEvent(&Event{n.ID, LEAVE, n.Addr})
				timeouted = append(timeouted, n.ID)
			}
		default:
			if now-n.updateAt > NODE_TIMEOUT {
				glog.Infof("Node:%s timeouted", n.ID)
				s.AddExchangeEvent(&Event{n.ID, LEAVE, n.Addr})
				timeouted = append(timeouted, n.ID)
			}
		}
		n.Unlock()
	}
	slice.RUnlock()

	for _, id := range timeouted {
		slice.Delete(id)
	}
}

func (s *Service) checkOuterSlice(slice *Slice, now int64) {

	var leader *Node
	if leader = slice.Leader(); leader == nil {
		return
	}

	leader.Lock()
	defer leader.Unlock()

	if now-leader.updateAt > SLICE_LEADER_TIMEOUT {
		s.route.Delete(leader.ID)
		s.notifyLock.Lock()
		s.notifyEvent = append(s.notifyEvent, &Event{leader.ID, LEAVE, leader.Addr})
		s.notifyLock.Unlock()
	}

}

func (s *Service) check() {

	now := time.Now().Unix()
	s.checkSelfSlice(s.selfSlice, now)

	if l := s.selfSlice.Leader(); l != nil && l.ID == s.id {
		for _, slice := range s.route.slices {
			if slice == s.selfSlice {
				s.checkSelfSlice(slice, now)
			} else {
				s.checkOuterSlice(slice, now)
			}
		}
		s.exchange()
		s.keepOtherAlive()

	} else {
		msg := new(Msg)
		msg.Type = EVENT_NOTIFICATION
		msg.From = s.id
		msg.Events = s.exchangeEvent
		msg.Events = append(msg.Events, ALIVE_EVENT)
		msg.Time = now
		s.SendMsg(l.Addr, msg)
	}

	s.exchangeLock.Lock()
	defer s.exchangeLock.Unlock()
	s.exchangeEvent = s.exchangeEvent[:0]

	s.notifyLock.Lock()
	defer s.notifyLock.Unlock()
	s.notifyEvent = s.notifyEvent[:0]
}

func (s *Service) AddNotifyEvent(msg *Msg) {
	s.notifyLock.Lock()
	defer s.notifyLock.Unlock()
	s.notifyEvent = append(s.notifyEvent, msg.Events...)
}
func (s *Service) AddExchangeMsg(msg *Msg) {

	s.exchangeLock.Lock()
	defer s.exchangeLock.Unlock()
	s.exchangeEvent = append(s.exchangeEvent, msg.Events...)
}
func (s *Service) AddExchangeEvent(event *Event) {

	s.exchangeLock.Lock()
	defer s.exchangeLock.Unlock()
	s.exchangeEvent = append(s.exchangeEvent, event)
}
