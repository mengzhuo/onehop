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
		n := s.route.GetNode(msg.From)
		if n != nil {
			n.updateAt = time.Now()
		}

	case MESSAGE_EXCHANGE:
		s.DoMessageExchange(raddr, msg)
	case EVENT_NOTIFICATION:
		s.DoEventNotification(raddr, msg)
	default:
		glog.Infof("UnKnown message type %v", msg)
		return
	}
}

func (s *Service) DoEventNotification(raddr *net.UDPAddr, msg *Msg) {
	// Any form of notify MUST be deal with in T(min)
	s.eventNotify = append(s.eventNotify, msg)

	if msg.Events == nil {
		// This node will take over our leadership
		slice := s.getMySlice()
		s.tick(slice)
		n := slice.Get(msg.From)
		if n == nil {
			n = &Node{msg.From, raddr, time.Now()}
			s.route.Add(n)
		}
		n.updateAt = time.Now()
		return
	}

	s.handleEvents(msg)
}

func (s *Service) DoMessageExchange(raddr *net.UDPAddr, msg *Msg) {
	s.exchangeMsg = append(s.exchangeMsg, msg)
	s.handleEvents(msg)
}

// Exchange with other slice leader
func (s *Service) Exchange(msg *Msg) {

	for _, slice := range s.route.slices {

		if slice.Leader == nil {
			continue
		}
		if slice.Leader.ID.Cmp(s.id) == 0 {
			// It's ourself
			continue
		}
		glog.V(9).Infof("Exchange with :%x", slice.Leader.ID)
		s.SendMsg(slice.Leader.Addr, msg)
	}
}

// The serivce should take whatever remote sends from bootstrap stage
// even it's a timeouted event
// Then send to its/other slice master about it's join event
func (s *Service) BootStrapReponse(raddr *net.UDPAddr, msg *Msg) {

	// Try with old slice leader
	slice_idx := s.route.GetIndex(s.id)
	glog.Infof("BootStrap response %s", msg)

	old_leader := false
	for _, rn := range msg.Events {
		n := rn.ToNode()
		s.route.Add(n)
		nslice_idx := s.route.GetIndex(n.ID)

		if nslice_idx != slice_idx {
			continue
		}

		old_leader = true
		nslice := s.route.slices[nslice_idx]

		nmsg := new(Msg)
		nmsg.NewID()
		nmsg.Type = EVENT_NOTIFICATION
		nmsg.From = s.id

		if nslice.Leader.ID.Cmp(s.id) == 0 {
			// We became leader notify leader to tick
			nmsg.Events = nil
		} else {
			e := Event{s.ID(), time.Now(), JOIN, s.conn.LocalAddr().String()}
			nmsg.Events = []Event{e}
		}
		glog.V(3).Infof("Send Event %s", nmsg)
		s.SendMsg(n.Addr, nmsg)
	}

	// we are new leader...tell other slice leader
	if !old_leader {
		glog.V(2).Infof("%x are new leader ", s.id)
		msg.Type = MESSAGE_EXCHANGE
		msg.NewID()
		event := &Event{s.ID(), time.Now(), JOIN,
			s.conn.LocalAddr().String()}
		msg.Events = []Event{*event}
		s.Exchange(msg)
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

	if msg.Events == nil {
		return
	}

	for _, e := range msg.Events {

		if e.Time.Add(EVENT_TIMEOUT).Before(time.Now()) {
			glog.Infof("Recv timeouted event:%v", e)
			continue
		}
		if e.ID.Cmp(s.id) == 0 {
			continue
		}

		switch e.Status {
		case JOIN:
			if n := s.route.GetNode(e.ID); n != nil {
				n.updateAt = time.Now()
			} else {
				s.route.Add(e.ToNode())
			}
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
		s.route.Add(n)
	}

	n.updateAt = time.Now()
	s.handleEvents(msg)

	slice := s.getMySlice()
	idx := slice.getID(s.id)

	// Response to requester
	responseMsg := new(Msg)
	responseMsg.ID = msg.ID
	responseMsg.From = s.id
	responseMsg.Events = msg.Events
	responseMsg.Type = KEEP_ALIVE_RESPONSE

	s.SendMsg(raddr, responseMsg)

	// how can we pass msg on next node?
	cmp := msg.From.Cmp(s.id)

	var next *Node
	switch {
	case cmp == -1 && idx+1 < slice.Len():
		next = slice.nodes[idx+1]
	case cmp == 1 && idx > 0:
		// msg bigger than us msg to smaller node
		next = slice.nodes[idx-1]
	}
	if next != nil {
		msg.From = s.ID()
		s.SendTimeoutMsg(next.Addr, next, msg)
	}
}
