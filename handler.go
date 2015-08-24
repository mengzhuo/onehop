package onehop

import (
	"net"
	"sync"
	"time"

	"github.com/golang/glog"
)

const (
	SLICE_LEADER_TIMEOUT = 5
	EVENT_TIMEOUT        = 10
	NODE_TIMEOUT         = 10
)

func (s *Service) handle(raddr *net.UDPAddr, msg *Msg) {

	switch msg.Type {

	case BOOTSTRAP:

		msg.Type = BOOTSTRAP_RESPONSE
		msg.Time = time.Now().Unix()
		msg.From = s.id
		for _, slice := range s.route.slices {
			slice.RLock()
			for _, n := range slice.Nodes {
				msg.Events = append(msg.Events, &Event{n.ID, JOIN, n.Addr})
			}
			slice.RUnlock()
		}
		s.SendMsg(raddr, msg)

	case BOOTSTRAP_RESPONSE:
		s.OnBootstrapResponse(raddr, msg)

	case KEEP_ALIVE:
		s.OnKeepAlive(raddr, msg)
		s.AddExchangeMsg(msg)

	case KEEP_ALIVE_RESPONSE:
		s.handleEvents(raddr, msg)
		s.AddExchangeMsg(msg)

	case MESSAGE_EXCHANGE:
		s.handleEvents(raddr, msg)
		s.AddNotifyEvent(msg)

	case EVENT_NOTIFICATION:
		s.handleEvents(raddr, msg)
		s.AddExchangeMsg(msg)
	default:
		glog.Infof("UnKnown message type %v", msg)
		return
	}
}

func (s *Service) handleEvents(raddr *net.UDPAddr, msg *Msg) {

	slice := s.route.GetSlice(msg.From)

	if n := slice.Get(msg.From); n == nil {
		glog.Infof("Msg Add Node:%s", msg.From)
		s.route.Add(&Node{msg.From, raddr, msg.Time, &sync.Mutex{}})
	} else {
		n.Update(msg.Time)
	}
	if msg.Events == nil {
		return
	}

	for _, event := range msg.Events {

		if event.ID == s.id {
			continue
		}

		switch event.Status {
		case JOIN:
			id_slice := s.route.GetSlice(event.ID)
			if n := id_slice.Get(event.ID); n != nil && n.updateAt < msg.Time {
				n.Update(msg.Time)
				continue
			}
			s.route.Add(&Node{event.ID, event.Addr, msg.Time, &sync.Mutex{}})
		case LEAVE:
			s.route.Delete(event.ID)
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
		msg.Events = append(msg.Events, &Event{n.ID, JOIN, s.conn.LocalAddr().(*net.UDPAddr)})
		msg.Time = time.Now().Unix()
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

func (s *Service) keepOtherAlive() {

	msg := new(Msg)
	msg.From = s.id
	msg.Type = KEEP_ALIVE
	msg.Events = append(s.exchangeEvent, s.notifyEvent...)
	msg.Time = time.Now().Unix()

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

	s.exchangeLock.RLock()
	defer s.exchangeLock.RUnlock()
	now := time.Now().Unix()

	for _, slice := range s.route.slices {
		if slice == s.selfSlice {
			continue
		}

		if n := slice.Leader(); n != nil {

			msg := &Msg{Type: MESSAGE_EXCHANGE,
				From:   s.id,
				Events: s.exchangeEvent,
				Time:   now}
			s.SendMsg(n.Addr, msg)
		}
	}
}
