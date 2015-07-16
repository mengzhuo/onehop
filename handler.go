package onehop

import (
	"net"
	"time"

	"github.com/golang/glog"
)

const EVENT_TIMEOUT = 3 * time.Second

func (s *Service) checkResponse(id uint32) (existed bool) {

	s.replyLock.Lock()
	defer s.replyLock.Unlock()

	if _, existed = s.requests[id]; !existed {
		glog.V(3).Infof("Msg ID:%d Type:%d timeouted/not in list", id)
		return
	}
	delete(s.requests, id)
	return existed
}

func (s *Service) Handle(raddr *net.UDPAddr, msg *Msg) {

	// Main handle function for ALL msg
	switch msg.Type {

	case BOOTSTRAP:
		s.BootStrap(raddr, msg)
	case BOOTSTRAP_RESPONSE:
		s.BootStrapReponse(raddr, msg)

	case KEEP_ALIVE:
		s.KeepAlive(raddr, msg)
	case KEEP_ALIVE_RESPONSE:
		if !s.checkResponse(msg.ID) {
			return
		}
		s.route.Refresh(msg.From)

	case MESSAGE_EXCHANGE:
		s.DoMessageExchange(msg)
	case EVENT_NOTIFICATION:
		s.DoEventNotification(msg)
	default:
		glog.Infof("UnKnown message type %v", msg)
		return
	}
}

func (s *Service) DoEventNotification(msg *Msg) {
	// Any form of notify MUST be deal with in T(min)
	s.eventNotify = append(s.eventNotify, msg)
}

func (s *Service) DoMessageExchange(msg *Msg) {
	s.exchangeMsg = append(s.exchangeMsg, msg)
}

// Exchange with other slice leader
func (s *Service) Exchange(msg *Msg) {

	saddr := s.conn.LocalAddr().String()

	for _, slice := range s.route.slices {

		if slice.Leader == nil {
			continue
		}
		if slice.Leader.Addr.String() == saddr {
			continue
		}
		s.SendMsg(slice.Leader.Addr, msg)
	}
}

// The serivce should take whatever remote sends from bootstrap stage
// even it's a timeouted event
// Then send to its/other slice master about it's join event
func (s *Service) BootStrapReponse(raddr *net.UDPAddr, msg *Msg) {

	// Try with old slice leader
	slice_idx, _ := s.route.GetIndex(s.id)
	slice := s.route.slices[slice_idx]

	old_leader := slice.Leader

	for _, rn := range msg.Events {
		n := rn.ToNode()
		s.route.Add(n)
		n.resetTimer()
	}

	// we are new leader...tell other slice leader
	if old_leader == nil || slice.Leader.ID == s.id {
		msg.Type = MESSAGE_EXCHANGE
		msg.NewID()
		event := &Event{s.ID(), time.Now(), JOIN,
			s.conn.LocalAddr().String()}
		msg.Events = []Event{*event}
		s.Exchange(msg)
	}

	if old_leader != nil {
		//Notify our leader
		msg.NewID()
		msg.Type = EVENT_NOTIFICATION
		s.SendMsg(old_leader.Addr, msg)
	}
}

// With bootstrap we should return all slice leaders,
// Let requester decide which leader should talk to by
// itself
func (s *Service) BootStrap(raddr *net.UDPAddr, msg *Msg) {

	events := make([]Event, 0)
	for _, s := range s.route.slices {
		if s.Leader == nil {
			continue
		}
		n := s.Leader
		e := &Event{n.ID, time.Now(), JOIN, n.Addr.String()}
		events = append(events, *e)
	}
	msg.Events = events
	msg.From = s.ID()
	msg.Type = BOOTSTRAP_RESPONSE
	s.SendMsg(raddr, msg)
}

func (s *Service) handleEvents(msg *Msg) {

	for _, e := range msg.Events {

		if e.Time.Add(EVENT_TIMEOUT).Before(time.Now()) {
			glog.Infof("Recv timeouted event:%v", e)
			continue
		}
		switch e.Status {
		case JOIN:
			if n := s.route.GetNode(e.ID); n != nil {
				n.updateAt = time.Now()
			}
			s.route.Add(e.ToNode())
		case LEAVE:
			s.route.Delete(e.ID)
		}
	}

}

func (s *Service) KeepAlive(raddr *net.UDPAddr, msg *Msg) {

	glog.V(4).Infof("KeepAlive Msg:%v", msg)
	n := s.route.GetNode(msg.From)
	if n == nil {
		n = &Node{ID: msg.From, Addr: raddr, updateAt: time.Now()}
	}
	s.route.Add(n)
	s.route.Refresh(n.ID)

	_, unit := s.getMySliceUnit()
	idx := unit.getID(s.id)
	// how can we pass msg on next node?
	cmp := msg.From.Cmp(s.id)

	var next *Node
	switch {
	case cmp == -1 && idx+1 < unit.Len():
		next = unit.nodes[idx+1]
	case cmp == 1 && idx > 0:
		// msg bigger than us msg to smaller node
		next = unit.nodes[idx-1]
	}
	msg.From = s.ID()
	msg.NewID()
	s.SendTimeoutMsg(next.Addr, next, msg)

	// Response to requester
	responseMsg := new(Msg)
	responseMsg.ID = msg.ID
	responseMsg.From = msg.From
	responseMsg.Events = msg.Events
	responseMsg.Type = KEEP_ALIVE_RESPONSE

	s.SendMsg(raddr, responseMsg)

}
