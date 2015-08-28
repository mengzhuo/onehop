package onehop

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/rpc"
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

	RPCPool *RPCPool
	DB      *Storage
	Booted  chan bool
}

func NewUDPDecoder(buf *bytes.Buffer) *json.Decoder {
	return json.NewDecoder(buf)
}

// NetType, Address for UDP connection
// k for OneHop slice number
func NewService(netType, address string, k, w, r int) *Service {

	glog.Info("Starting OneHop K/V service")
	listener, err := newListener(netType, address)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	glog.Infof("Listening on:%s %s", netType, address)
	glog.Infof("Config Write:%d, Read:%d", k, r)
	route = NewRoute(k)
	vid := make([]byte, 16)
	rand.Read(vid)
	id := hex.EncodeToString(vid)

	slice := route.slices[route.GetIndex(id)]
	glog.Infof("Node ID:%s", id)

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
		NewRPCPool(),
		NewStorage(),
		make(chan bool),
	}

	if err = rpc.Register(service.DB); err != nil {
		glog.Fatal(err)
	}
	if rpc_listener, err := net.Listen("tcp", listener.LocalAddr().String()); err != nil {
		glog.Fatal(err)
	} else {
		go rpc.Accept(rpc_listener)
	}

	glog.V(3).Infof("RPC ..... [on]")

	ALIVE_EVENT = &Event{id, JOIN, listener.LocalAddr().(*net.UDPAddr)}
	return service
}

func (s *Service) BootStrap(address string) {

	glog.Infof("BootStrap From ..... %s", address)
	addr, _ := net.ResolveUDPAddr("udp", address)
	msg := new(Msg)
	msg.Type = BOOTSTRAP
	msg.From = s.id
	s.SendMsg(addr, msg)

}

func (s *Service) Replicate() {

	id := s.id
	<-s.Booted
	glog.Infof("Start replication")
	for i := 0; i < s.R; i++ {
		data := make(map[string]*Item, 0)
		node := s.route.SuccessorOf(id)
		if node == nil || node.ID == s.id {
			break
		}
		c, err := s.RPCPool.Get(node.Addr.String())
		if err != nil {
			glog.Error(err)
			s.RPCError(err, node)
		}
		c.Call("Storage.Replicate", s.id, &data)
		glog.Infof("Get %d items from Node:%s", len(data), node.ID)
		for k, v := range data {
			s.DB.mu.Lock()
			s.DB.db[k] = v
			s.DB.mu.Unlock()
		}
		id = node.ID
	}
}

func (s *Service) BootStrapAndServe(address string) {

	go s.listen()
	s.addSelfNode()
	glog.Infof("Serve Mode: BootStraped")
	s.BootStrap(address)
	go s.Replicate()
	s.startRouteChecker()
}

func (s *Service) StandaloneServe() {
	go s.listen()
	s.addSelfNode()
	glog.Infof("Serve Mode: Stand Alone")
	s.startRouteChecker()
}

func BytesToId(p []byte) string {
	return fmt.Sprintf("%032x", p)
}

func (s *Service) listen() {

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

	if len(msg.Events) <= MAX_EVENT_PER_MSG {
		p, err := json.Marshal(msg)
		if err != nil {
			glog.Error(err)
			return
		}
		glog.V(10).Infof("SEND %s with %x", dstAddr, p)
		s.Send(dstAddr, p)
		return
	}

	events := msg.Events[:]

	for i := 0; len(events) > 0; i++ {

		if len(events) < MAX_EVENT_PER_MSG {
			msg.Events = events[:]
		} else {
			msg.Events = events[:MAX_EVENT_PER_MSG]
		}

		p, err := json.Marshal(msg)
		if err != nil {
			glog.Error(err)
			return
		}
		glog.V(10).Infof("SEND splited %s with %x", dstAddr, p)
		s.Send(dstAddr, p)

		if len(events) < MAX_EVENT_PER_MSG {
			events = events[:0]
		} else {
			events = events[MAX_EVENT_PER_MSG:]
		}
	}
}

func (s *Service) addSelfNode() {

	n := &Node{s.id,
		s.conn.LocalAddr().(*net.UDPAddr),
		time.Now().Unix(),
		&sync.Mutex{}}
	route.Add(n)
}

func (s *Service) startRouteChecker() {

	glog.Info("Route checker....[on]")

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

	} else if now%3 == 0 {
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
