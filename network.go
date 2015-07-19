// Package onehop provides ...
package onehop

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net"
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
	id       *big.Int

	Ticker *time.Ticker

	exchangeMsg []*Msg
	eventNotify []*Msg

	selfSlice *Slice
	pinger    *Node
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
func NewService(netType, address string, k int) *Service {

	listener, err := newListener(netType, address)
	if err != nil {
		panic(err)
	}

	glog.Infof("Listening to :%s %s", netType, address)
	route = NewRoute(k)
	bp := &BytePool{make(chan []byte, 1024), 8192}
	max := new(big.Int).SetBytes(FullID)

	id, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	glog.Infof("initial id:%x", id)

	n := &Node{ID: id, Addr: listener.LocalAddr().(*net.UDPAddr)}
	route.Add(n)

	eventTicker := time.NewTicker(3 * time.Second)
	slice := route.slices[route.GetIndex(id)]

	service = &Service{listener, route, bp, id,
		eventTicker,
		make([]*Msg, 0), make([]*Msg, 0), slice, n}

	go service.Tick()
	go route.ServeTimeout(slice)
	return service
}

func (s *Service) ID() *big.Int {
	return new(big.Int).SetBytes(s.id.Bytes())
}

func (s *Service) IAmSliceLeader() (answer bool) {

	if s.selfSlice.Leader != nil && s.selfSlice.Leader.ID.Cmp(s.id) == 0 {
		answer = true
	}
	return
}

func (s *Service) MsgToEvents(msgs []*Msg) []Event {

	es := make([]Event, 0)
	for _, m := range msgs {

		for _, e := range m.Events {
			if e.Time.Add(EVENT_TIMEOUT).Before(time.Now()) {
				continue
			}
			es = append(es, e)
		}
		es = append(es, m.Events...)
	}
	return es
}

func (s *Service) getMySlice() (slice *Slice) {

	slice = s.route.slices[s.route.GetIndex(s.id)]
	return
}

func (s *Service) Tick() {
	slice := s.getMySlice()
	for {
		select {
		case _ = <-s.Ticker.C:
			s.tick(slice)
		case n := <-s.route.timeOutNode:
			s.NotifySliceLeader(n, LEAVE)
		}
	}

}

func (s *Service) tick(slice *Slice) {

	if s.IAmSliceLeader() {
		// We are slice leader
		// put all event_notify to unit leader
		// And exchange with other slice leader
		emsg := NewMsg(MESSAGE_EXCHANGE, s.ID(),
			s.MsgToEvents(s.eventNotify))
		emsg.Events = append(emsg.Events,
			Event{s.id, time.Now(), JOIN, s.conn.LocalAddr().String()})
		s.Exchange(emsg)

		emsg.Events = append(emsg.Events, s.MsgToEvents(s.exchangeMsg)...)
		msg := NewMsg(KEEP_ALIVE,
			s.ID(), emsg.Events)

		n := slice.successorOf(s.id)
		if n != nil {
			s.SendMsg(n.Addr, msg)
		}
		pn := slice.predecessorOf(s.id)
		if pn != nil {
			s.SendMsg(pn.Addr, msg)
		}

	}

	if s.pinger != nil && s.pinger.Addr != s.conn.LocalAddr() {
		if s.pinger.updateAt.Add(NODE_TIMEOUT).Before(time.Now()) {
			// Timeouted
			s.route.Delete(s.pinger.ID)
			s.NotifySliceLeader(s.pinger, LEAVE)
		}
	}

	// Reset all events
	s.exchangeMsg = s.exchangeMsg[:0]
	s.eventNotify = s.eventNotify[:0]
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
		glog.V(10).Infof("Recv From:%x TYPE:%x", msg.From, msg.Type)
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

	slice := s.getMySlice()

	if slice.Leader == nil {
		glog.Errorf("Something is wrong!!! we are in the slice however there is no slice leader?")
		return
	}

	msg := new(Msg)
	msg.NewID()
	msg.From = s.id
	e := Event{n.ID, time.Now(), status, n.Addr.String()}
	msg.Events = append(msg.Events, e)

	if s.IAmSliceLeader() {
		// we are leader now
		msg.Events = append(msg.Events, Event{s.id, time.Now(), JOIN, s.conn.LocalAddr().String()})
		msg.Type = MESSAGE_EXCHANGE
		s.Exchange(msg)
		return
	}

	msg.Type = EVENT_NOTIFICATION
	s.SendMsg(slice.Leader.Addr, msg)
}
