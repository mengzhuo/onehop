package onehop

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"strings"
	"time"

	"github.com/golang/glog"
)

type Service struct {
	conn *net.UDPConn

	route   *Route
	counter uint8

	db       *Storage
	bytePool *BytePool
	id       *big.Int
	selfNode *Node

	ticker *time.Ticker

	exchangeEvent []Event
	notifyEvent   []Event

	selfSlice   *Slice
	pinger      *Node
	leftPonger  *Node
	rightPonger *Node
	W, R        int
	RPCPool     *RPCPool
}

// NetType, Address for UDP connection
// k for OneHop slice number
func NewService(netType, address string, k, w, r int) *Service {
	glog.Infof("Write:%d, Read:%d", k, r)
	listener, err := newListener(netType, address)
	if err != nil {
		panic(err)
	}
	port := strings.SplitN(address, ":", 2)
	rpc_listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port[1]))
	if err != nil {
		panic(err)
	}

	glog.Infof("Listening to :%s %s", netType, address)
	route = NewRoute(k)
	bp := &BytePool{make(chan []byte, 1024), 16 * 1024}
	max := new(big.Int).SetBytes(FullID)

	id, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	glog.Infof("initial id:%x", id)

	n := &Node{ID: id, Addr: listener.LocalAddr().(*net.UDPAddr)}
	route.Add(n)

	eventTicker := time.NewTicker(1 * time.Second)
	slice := route.slices[route.GetIndex(id)]

	service = &Service{listener, route, uint8(0), NewStorage(),
		bp, id, n,
		eventTicker,
		make([]Event, 0), make([]Event, 0), slice, n,
		nil, nil, w, r, &RPCPool{make(map[string]*rpc.Client, 0)}}

	rpc.Register(service.db)
	rpc.HandleHTTP()
	go http.Serve(rpc_listener, nil)
	glog.V(3).Infof("RPC Listener Accepted")
	return service
}

func (s *Service) GetByString(key string) *Item {

	k := []byte(key)
	return s.Get(k)

}

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

func (s *Service) ID() *big.Int {
	return new(big.Int).SetBytes(s.id.Bytes())
}

func (s *Service) IAmSliceLeader() (answer bool) {

	// s.selfSlice leader should not be nil
	// it's fair that raise a panic if this happen

	if s.selfSlice.Leader == s.selfNode {
		answer = true
	}
	return
}

func (s *Service) Tick() {

	for {
		select {
		case _ = <-s.ticker.C:
			s.tick()
		}
	}

}

func (s *Service) BootStrapFrom(address string) {

	glog.Infof("BootStrap From :%s", address)
	addr, _ := net.ResolveUDPAddr("udp", address)
	msg := &Msg{}
	msg.Type = BOOTSTRAP
	msg.NewID()
	s.SendMsg(addr, msg)
}

func BytesToId(p []byte) (b *big.Int) {
	b = new(big.Int)
	b.SetBytes(p)
	return
}

func (s *Service) tick() {

	now := time.Now()
	s.selfNode.updateAt = now

	// Check all slice leaders
	for _, slice := range s.route.slices {

		if slice.Leader != nil && slice.Leader != s.selfNode &&
			time.Since(slice.Leader.updateAt).Seconds() > SLICE_LEADER_TIMEOUT {
			// Timeouted
			glog.Errorf("Slice Leader %s timeout", slice.Leader)
			s.NotifySliceLeader(slice.Leader, LEAVE)
			s.route.Delete(slice.Leader.ID)
			// Give new leader a chance
			if slice.Leader != nil {
				slice.Leader.updateAt = now
			}
		}
	}

	if s.pinger != nil && s.pinger != s.selfNode &&
		time.Since(s.pinger.updateAt).Seconds() > NODE_TIMEOUT {
		// Timeouted
		glog.Errorf("Pinger %s timeout", s.pinger)
		s.NotifySliceLeader(s.pinger, LEAVE)
		s.route.Delete(s.pinger.ID)
		s.pinger = nil
	}

	if s.leftPonger != nil && s.leftPonger != s.selfNode &&
		time.Since(s.leftPonger.updateAt).Seconds() > NODE_TIMEOUT {
		// Timeouted
		glog.Errorf("Left Ponger %s timeout", s.leftPonger)
		s.NotifySliceLeader(s.leftPonger, LEAVE)
		s.route.Delete(s.leftPonger.ID)
		s.leftPonger = nil
	}

	if s.rightPonger != nil && s.rightPonger != s.selfNode &&
		time.Since(s.rightPonger.updateAt).Seconds() > NODE_TIMEOUT {
		// Timeouted
		glog.Errorf("Right Ponger %s timeout", s.rightPonger)
		s.NotifySliceLeader(s.rightPonger, LEAVE)
		s.route.Delete(s.rightPonger.ID)
		s.rightPonger = nil
	}

	if s.IAmSliceLeader() {
		// We are slice leader
		// put all event_notify to unit leader
		// And exchange with other slice leader
		emsg := NewMsg(MESSAGE_EXCHANGE, s.ID(), s.notifyEvent)
		emsg.Events = append(emsg.Events, Event{s.id, now, JOIN, s.conn.LocalAddr().String()})

		// Each 21 seconds notify other slice leader about all our nodes
		if s.counter%21 == 0 {

			to_delete := make([]*Node, 0)
			for _, n := range s.selfSlice.nodes {

				status := JOIN

				if time.Since(n.updateAt).Seconds() > NODE_TIMEOUT {
					status = LEAVE
					to_delete = append(to_delete, n)
				}

				emsg.Events = append(emsg.Events,
					Event{n.ID, now, status, n.Addr.String()})
			}

			for _, n := range to_delete {
				s.route.Delete(n.ID)
			}
		}

		s.Exchange(emsg)

		emsg.Events = append(emsg.Events, s.exchangeEvent...)
		msg := NewMsg(KEEP_ALIVE, s.ID(), emsg.Events)

		n := s.selfSlice.successorOf(s.id)
		if n != nil {
			s.SendMsg(n.Addr, msg)
			s.leftPonger = n
		}

		pn := s.selfSlice.predecessorOf(s.id)
		if pn != nil && pn != n {
			s.SendMsg(pn.Addr, msg)
			s.rightPonger = pn
		}

	} else {
		// Each 13 seconds, tell our leader our existence
		if s.selfSlice.Leader != s.pinger &&
			s.counter%(10) == 0 {
			msg := NewMsg(EVENT_NOTIFICATION, s.id,
				[]Event{Event{s.id, now, JOIN,
					s.conn.LocalAddr().String()}})
			s.SendMsg(s.selfSlice.Leader.Addr, msg)
		}
	}

	// Reset all events
	s.exchangeEvent = s.exchangeEvent[:0]
	s.notifyEvent = s.notifyEvent[:0]
	s.counter += 1
}

func (s *Service) Listen() {

	go service.Tick()

	for {

		p := s.bytePool.Get()
		n, addr, err := s.conn.ReadFromUDP(p)
		if err != nil || n < 3 {
			glog.Errorf("insufficient data for parsing...", addr, p[:n])
			s.bytePool.Put(p)
			continue
		}

		msg := new(Msg)
		err = json.Unmarshal(p[:n], msg)
		if err != nil {
			panic(err)
		}
		glog.V(10).Infof("Recv From:%x TYPE:%s", msg.From, typeName[msg.Type])
		s.Handle(addr, msg)
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
		glog.Errorf("Error on parse %s", msg)
		return
	}
	s.Send(dstAddr, p)
}

func (s *Service) NotifySliceLeader(n *Node, status byte) {

	slice := s.selfSlice

	if slice.Leader == nil {
		glog.Fatal("Something is wrong!!! we are still in the slice and there is no slice leader?")
	}
	glog.V(5).Infof("Notify Slice leader about  %s", n)
	e := Event{n.ID, time.Now(), status, n.Addr.String()}

	s.exchangeEvent = append(s.exchangeEvent, e)

	if !s.IAmSliceLeader() {
		// we are leader now
		msg := new(Msg)
		msg.NewID()
		msg.From = s.id
		msg.Events = append(msg.Events, e)
		msg.Type = EVENT_NOTIFICATION
		s.SendMsg(slice.Leader.Addr, msg)
	}
}
