package onehop

import (
	"net"
	"time"

	"github.com/golang/glog"
)

const (
	SLICE_LEADER_TIMEOUT = 15
	EVENT_TIMEOUT        = 20
	NODE_TIMEOUT         = 30
)

func (s *Service) handle(raddr *net.UDPAddr, msg *Msg) {

	switch msg.Type {

	case BOOTSTRAP:
		msg.Events = make(map[string]*Event)
		msg.Type = BOOTSTRAP_RESPONSE
		msg.Time = time.Now()
		msg.From = s.id
		for _, slice := range s.route.slices {
			slice.RLock()
			for _, n := range slice.Nodes {
				msg.Events[n.ID] = &Event{JOIN, n.Addr}
			}
			slice.RUnlock()
		}
		s.SendMsg(raddr, msg)

	case BOOTSTRAP_RESPONSE:
		s.OnBootstrapResponse(raddr, msg)
	case KEEP_ALIVE:
		s.OnKeepAlive(raddr, msg)
	case KEEP_ALIVE_RESPONSE:
		s.handleEvents(raddr, msg)
	case MESSAGE_EXCHANGE:
		s.handleEvents(raddr, msg)
	case EVENT_NOTIFICATION:
	default:
		glog.Infof("UnKnown message type %v", msg)
		return
	}
}

func (s *Service) handleEvents(raddr *net.UDPAddr, msg *Msg) {

	if n := s.route.GetNode(msg.From); n == nil {
		glog.Infof("Msg Add Node:%s", msg.From)
		s.route.Add(&Node{msg.From, raddr, msg.Time})
	} else {
		n.updateAt = msg.Time
	}
	if msg.Events == nil {
		return
	}
	for id, e := range msg.Events {
		switch e.Status {
		case JOIN:
			if n := s.route.GetNode(id); n != nil {
				n.updateAt = msg.Time
				continue
			}
			s.route.Add(&Node{id, e.Addr, msg.Time})
		case LEAVE:
			s.route.Delete(id)
		}
	}

}

func (s *Service) passOn(msg *Msg) {

	var n *Node
	if msg.From > s.id {
		n = s.selfSlice.predecessorOf(s.id)
	} else {
		n = s.selfSlice.successorOf(s.id)
	}

	if n != nil {
		msg.From = s.id
		msg.Events[s.id] = &Event{JOIN, s.conn.LocalAddr().(*net.UDPAddr)}
		msg.Time = time.Now()
		s.SendMsg(n.Addr, msg)
	}
}

func (s *Service) OnBootstrapResponse(raddr *net.UDPAddr, msg *Msg) {
	s.handleEvents(raddr, msg)
}

func (s *Service) OnKeepAlive(raddr *net.UDPAddr, msg *Msg) {
	s.handleEvents(raddr, msg)
	s.passOn(msg)

	msg.Events = nil

	msg.Type = KEEP_ALIVE_RESPONSE
	msg.From = s.id
	s.SendMsg(raddr, msg)
}

func (s *Service) keepOtherAlive(msg *Msg) {

	sn := s.selfSlice.successorOf(s.id)
	if sn != nil {
		s.SendMsg(sn.Addr, msg)
	}

	pn := s.selfSlice.predecessorOf(s.id)
	if pn != nil {
		s.SendMsg(pn.Addr, msg)
	}

}

func (s *Service) exchange() {

	for _, slice := range s.route.slices {
		if slice == s.selfSlice {
			continue
		}

		if n := slice.Leader(); n != nil {

			msg := &Msg{Type: MESSAGE_EXCHANGE,
				From:   s.id,
				Events: s.notifyEvent,
				Time:   time.Now()}
			s.SendMsg(n.Addr, msg)
		}
	}
}
