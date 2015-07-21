// Package onehop provides ...
package onehop

import (
	"crypto/rand"
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

var (
	route   *Route
	service *Service
)

func newListener(netType, address string) (conn *net.UDPConn, err error) {

	uaddr, err := net.ResolveUDPAddr(netType, address)
	if err != nil {
		return nil, err
	}

	conn, err = net.ListenUDP(netType, uaddr)
	if err != nil {
		return nil, err
	}

	return conn, err
}

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
	rpcPool     *RPCPool
}

type BytePool struct {
	freeList chan []byte
	maxSize  int
}

func (b *BytePool) Get() (p []byte) {
	select {
	case p = <-b.freeList:
	default:
		p = make([]byte, b.maxSize)

	}
	return
}

func (b *BytePool) Put(p []byte) {

	select {
	case b.freeList <- p:
	default:
	}
}

// NetType, Address for UDP connection
// k for OneHop slice number
func NewService(netType, address string, k, w, r int) *Service {

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

	rpc.HandleHTTP()
	rpc.Register(service.db)

	go http.Serve(rpc_listener, nil)
	go service.Tick()
	return service
}

func (s *Service) Get(key []byte) *Item {

	items := make([]*Item, s.R)
	id := new(big.Int).SetBytes(key)
	for i := 0; i < s.R; i++ {
		node := s.route.SuccessorOf(id)
		if node == nil {
			continue
		}
		client, err := s.rpcPool.Get(node.Addr.String())
		if err != nil {
			glog.Error(err)
		}
		var reply *Item
		err = client.Call("Get", string(key), reply)
		if err != nil {
			items[i] = reply
		}
	}

	for _, item := range items {
		return item
	}
	return nil
}

func (s *Service) ID() *big.Int {
	return new(big.Int).SetBytes(s.id.Bytes())
}

func (s *Service) IAmSliceLeader() (answer bool) {

	if s.selfSlice.Leader != nil && s.selfSlice.Leader == s.selfNode {
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

func (s *Service) tick() {

	now := time.Now()
	// After 128 second it will trigger all node to sync data
	for _, slice := range s.route.slices {

		if slice.Leader != nil && slice.Leader != s.selfNode &&
			slice.Leader.updateAt.Add(SLICE_LEADER_TIMEOUT).Before(now) {
			// Timeouted
			glog.Errorf("Slice Leader %s timeout", slice.Leader)
			s.NotifySliceLeader(slice.Leader, LEAVE)
			s.route.Delete(slice.Leader.ID)
		}
	}

	if s.pinger != nil && s.pinger != s.selfNode &&
		s.pinger.updateAt.Add(NODE_TIMEOUT).Before(now) {
		// Timeouted
		glog.Errorf("Pinger %s timeout", s.pinger)
		s.NotifySliceLeader(s.pinger, LEAVE)
		s.route.Delete(s.pinger.ID)
		s.pinger = nil
	}

	if s.leftPonger != nil && s.leftPonger != s.selfNode &&
		s.leftPonger.updateAt.Add(NODE_TIMEOUT).Before(now) {
		// Timeouted
		glog.Errorf("Left Ponger %s timeout", s.leftPonger)
		s.route.Delete(s.leftPonger.ID)
		s.NotifySliceLeader(s.leftPonger, LEAVE)

		s.leftPonger = nil
	}

	if s.rightPonger != nil && s.rightPonger != s.selfNode &&
		s.rightPonger.updateAt.Add(NODE_TIMEOUT).Before(now) {
		// Timeouted
		glog.Errorf("Right Ponger %s timeout", s.rightPonger)
		s.route.Delete(s.rightPonger.ID)
		s.NotifySliceLeader(s.rightPonger, LEAVE)
		s.rightPonger = nil
	}

	if s.IAmSliceLeader() {
		// We are slice leader
		// put all event_notify to unit leader
		// And exchange with other slice leader
		emsg := NewMsg(MESSAGE_EXCHANGE, s.ID(), s.notifyEvent)
		emsg.Events = append(emsg.Events, Event{s.id, now, JOIN, s.conn.LocalAddr().String()})

		// Each 22 seconds notify other slice leader about all nodes
		if s.counter%22 == 0 {
			glog.V(5).Info("Tell other leader about our slice")
			for _, n := range s.selfSlice.nodes {
				emsg.Events = append(emsg.Events, Event{n.ID, now, JOIN, n.Addr.String()})
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
		if pn != nil {
			s.SendMsg(pn.Addr, msg)
			s.rightPonger = n
		}

	} else {
		// Each 10 seconds, tell our leader
		if s.counter%13 == 0 {
			glog.V(5).Infof("Tell leader %x about ourself", s.selfSlice.Leader.ID)
			msg := NewMsg(EVENT_NOTIFICATION, s.id,
				[]Event{Event{s.id, now, JOIN, s.conn.LocalAddr().String()}})
			s.SendMsg(s.selfSlice.Leader.Addr, msg)
		}
	}

	// Reset all events
	s.exchangeEvent = s.exchangeEvent[:0]
	s.notifyEvent = s.notifyEvent[:0]
	s.counter += 1
}

func (s *Service) Listen() {

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
		glog.Errorf("Something is wrong!!! we are in the slice however there is no slice leader?")
		return
	}
	glog.V(5).Infof("Notify Slice leader about lossing %s", n)
	e := Event{n.ID, time.Now(), status, n.Addr.String()}

	if s.IAmSliceLeader() {
		// we are leader now
		s.exchangeEvent = append(s.exchangeEvent, e)
		return
	} else {
		msg := new(Msg)
		msg.NewID()
		msg.From = s.id
		msg.Events = append(msg.Events, e)
		msg.Type = EVENT_NOTIFICATION
		s.SendMsg(slice.Leader.Addr, msg)
	}
}
