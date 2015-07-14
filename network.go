// Package onehop provides ...
package onehop

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"net"
	"sync"
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
	conn  *net.UDPConn
	route *Route

	bytePool *BytePool
	msgPool  *MsgPool
	id       *big.Int

	requests       map[uint32]*Msg
	requestTimeout chan uint32
	replyLock      *sync.RWMutex

	Ticker        *time.Ticker
	pendingEvents []Event
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
// k, u for OneHop slice number and unit number
func NewService(netType, address string, k, u int) *Service {

	listener, err := newListener(netType, address)
	if err != nil {
		panic(err)
	}

	log.Printf("Listening to :%s %s", netType, address)
	route = NewRoute(k, u)
	bp := &BytePool{make(chan []byte, 1024), 8192}
	mp := &MsgPool{make(chan *Msg, 1024)}
	max := new(big.Int).SetBytes(FullID)

	id, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	glog.Infof("initial id:%x", id)

	n := &Node{ID: id, Addr: listener.LocalAddr().(*net.UDPAddr)}
	// route.Add(n)
	slice_idx, unit_idx := route.GetIndex(n.ID)
	slice := route.slices[slice_idx]
	unit := slice.units[unit_idx]
	result := unit.add(n)
	if !result {
		glog.Fatalf("Can not initialize node:%v", n)
	}
	slice.updateLeader()

	requests := make(map[uint32]*Msg, 0)
	requestTimeout := make(chan uint32, 1024)

	eventTicker := time.NewTicker(1 * time.Second)

	pendingEvents := make([]Event, 10)
	service = &Service{listener, route, bp, mp, id,
		requests, requestTimeout, &sync.RWMutex{},
		eventTicker, pendingEvents}
	return service
}

func (s *Service) Tick() {

	slice_idx, _ := s.route.GetIndex(s.id)
	self_slice := s.route.slices[slice_idx]

	for e := range s.Ticker.C {
		glog.V(8).Infof("ticker @ %s", e)
		if len(s.pendingEvents) == 0 || self_slice.Leader == nil {
			continue
		}
		msg := s.msgPool.Get()
		msg.From = s.id

		switch self_slice.Leader.ID.Cmp(s.id) {
		case -1, 1:
			//s.SendMsg
		case 0:
			// We are leader
			msg.Type = MESSAGE_EXCHANGE
			msg.Events = s.pendingEvents[:]
			s.Exchange(msg)
		}

		s.pendingEvents = s.pendingEvents[:0]
		s.msgPool.Put(msg)
	}
}

func (s *Service) Listen() {

	for {

		p := s.bytePool.Get()
		n, addr, err := s.conn.ReadFromUDP(p)
		glog.V(9).Infof("Recv data:%x", p[:n])
		if err != nil || n < 3 || p[0] != IDENTIFIER {
			glog.Errorf("insufficient data for parsing...", addr, p[:n])
			s.bytePool.Put(p)
			continue
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Error on get message %s", r)
				}
				s.bytePool.Put(p)
			}()
			msg := s.msgPool.Get()
			defer s.msgPool.Put(msg)
			err := json.Unmarshal(p[1:n], msg)
			if err != nil {
				panic(err)
			}
			s.Handle(addr, msg)
		}()
	}
}

func (s *Service) Send(dstAddr *net.UDPAddr, p []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Print(r)
		}
	}()

	buf := s.bytePool.Get()
	defer s.bytePool.Put(buf)

	buf[0] = IDENTIFIER
	buf = append(buf[:1], p...)
	s.conn.WriteToUDP(buf, dstAddr)
}

func (s *Service) SendMsg(dstAddr *net.UDPAddr, msg *Msg) {

	p, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error on parse %s", msg)
		return
	}
	s.Send(dstAddr, p)
}

func (s *Service) SendTimeoutMsg(dstAddr *net.UDPAddr, msg *Msg) {

	// it should not be any other msg id in here
	s.replyLock.Lock()
	defer s.replyLock.Unlock()
	s.requests[msg.ID] = msg
	id := msg.ID

	time.AfterFunc(NODE_TIMEOUT, func() {
		s.replyLock.Lock()
		defer s.replyLock.Unlock()
		log.Printf("Msg %x timeout checking...", msg.ID)
		if msg, ok := s.requests[id]; ok {
			delete(s.requests, id)
			log.Printf("Msg %d timeouted", msg.ID)
		}
	})
	s.SendMsg(dstAddr, msg)
}
