package onehop

import (
	"net"
	"time"

	"github.com/golang/glog"
)

const (
	SLICE_LEADER_TIMEOUT = 6 * time.Second
	EVENT_TIMEOUT        = 5 * time.Second
	NODE_TIMEOUT         = 10 * time.Second
)

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
		n := s.route.GetNode(msg.From)
		if n != nil {
			n.updateAt = time.Now()
		}

	case MESSAGE_EXCHANGE:
		s.DoMessageExchange(raddr, msg)
	case EVENT_NOTIFICATION:
		s.DoEventNotification(raddr, msg)
	case REPLICATE:
		s.Replicate(raddr, msg)
	case REPLICATE_RESPONSE:
		s.ReplicateResponse(raddr, msg)
	default:
		glog.Infof("UnKnown message type %v", msg)
		return
	}
}

func (s *Service) DoEventNotification(raddr *net.UDPAddr, msg *Msg) {
	// Any form of notify MUST be deal with in T(min)

	if msg.Events == nil {
		// This node will take over our leadership
		n := s.selfSlice.Get(msg.From)
		if n == nil {
			n = &Node{msg.From, raddr, time.Now()}
			s.route.Add(n)
		}
		n.updateAt = time.Now()
		s.exchangeEvent = append(s.exchangeEvent,
			Event{msg.From, time.Now(), JOIN, raddr.String()})
		s.tick()
		return
	} else {
		s.notifyEvent = append(s.notifyEvent, msg.Events...)
	}

	s.handleEvents(msg)
}

func (s *Service) DoMessageExchange(raddr *net.UDPAddr, msg *Msg) {
	if msg.Events == nil {
		return
	}
	s.exchangeEvent = append(s.exchangeEvent, msg.Events...)
	s.handleEvents(msg)
}

// Exchange with other slice leader
func (s *Service) Exchange(msg *Msg) {

	for _, slice := range s.route.slices {

		if slice.Leader == nil {
			continue
		}
		if slice == s.selfSlice {
			// It's ourself
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
	var old_leader *Node
	glog.Infof("BootStrap response %s", msg)

	for _, rn := range msg.Events {
		n := rn.ToNode()
		s.route.Add(n)

		nslice_idx := s.route.GetIndex(n.ID)
		if s.route.slices[nslice_idx] != s.selfSlice {
			continue
		}
		old_leader = n
	}

	// we are new leader...tell other slice leader
	if old_leader == nil {
		glog.V(2).Infof("%x are new leader ", s.id)
		msg.Type = MESSAGE_EXCHANGE
		msg.NewID()
		event := &Event{s.ID(), time.Now(), JOIN,
			s.conn.LocalAddr().String()}
		msg.Events = []Event{*event}
		s.Exchange(msg)
		return
	}

	msg.Type = EVENT_NOTIFICATION
	msg.From = s.id
	msg.NewID()
	if old_leader != s.selfSlice.Leader {
		// Tell old Leader about add
		msg.Events = nil
	} else {
		event := Event{s.ID(), time.Now(), JOIN,
			s.conn.LocalAddr().String()}
		msg.Events = []Event{event}
	}
	s.SendMsg(old_leader.Addr, msg)

	// Replicate from siblings

	for i := 0; i < s.R; i++ {
		node := s.route.SuccessorOf(s.id)
		if node != nil {
			s.goReplicate(node)
		}
	}
}

func (s *Service) goReplicate(node *Node) {
	client, err := s.rpcPool.Get(node.Addr.String())
	if err != nil {
		glog.Errorf("Replication failed on %x", node.ID)
	}
	var reply *map[string]*Item
	client.Call("Replicate", s.id.String(), reply)
	if reply == nil {
		glog.Errorf("Replication failed on %x return nil", node.ID)
		return
	}

	for k, v := range *reply {
		s.db.db[k] = v
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
		for _, n := range s.nodes {
			e := &Event{n.ID, time.Now(), JOIN, n.Addr.String()}
			events = append(events, *e)
		}
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

func (s *Service) Replicate(raddr *net.UDPAddr, msg *Msg) {

	glog.V(4).Infof("Recv Replicate:%s", msg)

	events := make([]Event, 0)

	for _, slice := range s.route.slices {

		for _, node := range slice.nodes {
			e := Event{node.ID, time.Now(), JOIN, node.Addr.String()}
			events = append(events, e)
		}
	}

	for i := 100; i < len(events); i += 100 {
		msg.Events = events[i-100 : i]
		msg.Type = REPLICATE_RESPONSE
		msg.From = s.ID()
		glog.V(9).Infof("Send replicate %s", msg)
		s.SendMsg(raddr, msg)
	}
}

func (s *Service) ReplicateResponse(raddr *net.UDPAddr, msg *Msg) {
	// we don't need to reponse with this
	glog.V(4).Infof("Recv Replicate Response:%s", msg)
	s.handleEvents(msg)
}

func (s *Service) KeepAlive(raddr *net.UDPAddr, msg *Msg) {

	n := s.route.GetNode(msg.From)
	if n == nil {
		n = &Node{ID: msg.From, Addr: raddr, updateAt: time.Now()}
		s.route.Add(n)
	}

	n.updateAt = time.Now()
	s.handleEvents(msg)

	slice := s.selfSlice
	idx := slice.getID(s.id)

	// Response to requester
	responseMsg := new(Msg)
	responseMsg.ID = msg.ID
	responseMsg.From = s.id
	responseMsg.Events = msg.Events
	responseMsg.Type = KEEP_ALIVE_RESPONSE

	s.pinger = n
	s.SendMsg(raddr, responseMsg)

	// how can we pass msg on next node?
	cmp := msg.From.Cmp(s.id)

	var next *Node
	switch {
	case cmp == -1 && idx+1 < slice.Len():
		next = slice.nodes[idx+1]
		s.rightPonger = next
		s.leftPonger = nil
	case cmp == 1 && idx > 0:
		// msg bigger than us msg to smaller node
		next = slice.nodes[idx-1]
		s.leftPonger = next
		s.rightPonger = nil
	default:
		// Nothin to pass on
		return
	}
	if next != nil {
		msg.From = s.ID()
		s.SendMsg(next.Addr, msg)
	}
}
