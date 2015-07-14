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
		glog.V(3).Infof("Msg ID:%d timeouted/not in list", id)
		return
	}
	delete(s.requests, id)
	return existed
}

func (s *Service) Handle(raddr *net.UDPAddr, msg *Msg) {

	// Main handle function for ALL msg
	switch msg.Type {

	case KEEP_ALIVE:
		s.KeepAlive(raddr, msg)
	case KEEP_ALIVE_RESPONSE:
		if !s.checkResponse(msg.ID) {
			return
		}
		s.route.Refresh(msg.From)

	case BOOTSTRAP:
		s.BootStrap(raddr, msg)
	case BOOTSTRAP_RESPONSE:
		if !s.checkResponse(msg.ID) {
			return
		}

	default:
		glog.Infof("UnKnown message type %v", msg)
		return
	}
}

// Exchange with other slice leader
func (s *Service) Exchange(msg *Msg) {

	slice_idx, _ := s.route.GetIndex(msg.From)

	for i, slice := range s.route.slices {
		if slice.Leader == nil || slice_idx == i {
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
	}

	// we are new leader...tell other slice leader
	if old_leader == nil || slice.Leader.ID == s.id {
		msg.Type = MESSAGE_EXCHANGE
		msg.NewID()
		event := &Event{s.id, time.Now(), JOIN,
			s.conn.LocalAddr().String()}
		msg.Events = []Event{*event}
		s.Exchange(msg)
	}

	if old_leader != nil {
		//Notify our leader
		msg.NewID()
		msg.Type = EVENT_NOTIFICATION
		msg.From = s.id
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
	msg.From = s.id
	msg.Type = BOOTSTRAP_RESPONSE
	s.SendMsg(raddr, msg)
}

func (s *Service) KeepAlive(raddr *net.UDPAddr, msg *Msg) {

	glog.V(4).Infof("KeepAlive Msg:%v", msg)
	n := s.route.GetNode(msg.From)
	if n == nil {
		n = &Node{ID: msg.From, Addr: raddr, updateAt: time.Now()}
	}
	s.route.Add(n)

	for _, rn := range msg.Events {

		if rn.Time.Add(EVENT_TIMEOUT).Before(time.Now()) {
			glog.Infof("Recv timeouted event:%v", rn)
			continue
		}
		n := rn.ToNode()
		switch rn.Status {
		case JOIN:
			s.route.Add(n)
		case LEAVE:
			s.route.Delete(n.ID)
		}
	}

	sidx, uidx := s.route.GetIndex(s.id)
	unit := s.route.slices[sidx].units[uidx]

	i := unit.getID(msg.From)
	cmp := msg.From.Cmp(s.id)

	msg.From = s.id
	switch {
	case cmp == -1 && i > 0:
		i--
	case cmp == 1 && (i < unit.Len()-1):
		i++
	default:
		// we are the end of Unit
		// Response to requester
		msg.Type = KEEP_ALIVE_RESPONSE
		msg.Events = nil
		s.SendMsg(raddr, msg)
		return
	}

	// Response to requester
	responseMsg := s.msgPool.Get()
	defer s.msgPool.Put(responseMsg)

	responseMsg.ID = msg.ID
	responseMsg.From = msg.From
	responseMsg.Events = msg.Events
	responseMsg.Type = KEEP_ALIVE_RESPONSE
	s.SendMsg(raddr, responseMsg)

	// pass msg to successor/predecessor in unit
	msg.NewID()
	nn := unit.nodes[i]
	s.route.Refresh(nn.ID)
	s.SendTimeoutMsg(nn.Addr, msg)
}
