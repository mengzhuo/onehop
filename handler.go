package onehop

import (
	"log"
	"net"
	"time"
)

const EVENT_TIMEOUT = 3 * time.Second

func (s *Service) Handle(raddr *net.UDPAddr, msg *Msg) {

	// Main handle function for ALL msg
	switch msg.Type {

	case KEEP_ALIVE:
		s.KeepAlive(raddr, msg)
	case KEEP_ALIVE_RESPONSE:
		s.route.Refresh(msg.From)
	default:
		log.Printf("UnKnown message type  %v", msg)
		return
	}
}

func (s *Service) KeepAlive(raddr *net.UDPAddr, msg *Msg) {

	n := s.route.GetNode(msg.From)
	if n == nil {
		n = &Node{ID: msg.From, Addr: raddr, updateAt: time.Now()}
	}
	s.route.Add(n)

	for _, rn := range msg.Events {

		if rn.Time.Add(EVENT_TIMEOUT).Before(time.Now()) {
			log.Printf("Recv timeouted event:%v", rn)
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
	// Send to successor/predecessor in unit
	msg.NewID()
	nn := unit.nodes[i]
	s.SendTimeoutMsg(nn.Addr, nn.ID, msg)
}
