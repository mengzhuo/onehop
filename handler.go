package onehop

import (
	"net"

	"github.com/golang/glog"
)

const (
	SLICE_LEADER_TIMEOUT = 15
	EVENT_TIMEOUT        = 20
	NODE_TIMEOUT         = 30
)

func (s *Service) Handle(raddr *net.UDPAddr, msg *MsgHeader, events []*Event) {

	switch msg.Type {

	case BOOTSTRAP:
	case BOOTSTRAP_RESPONSE:

	case KEEP_ALIVE:
	case KEEP_ALIVE_RESPONSE:
	case MESSAGE_EXCHANGE:
	case EVENT_NOTIFICATION:
	default:
		glog.Infof("UnKnown message type %v", msg)
		return
	}
}
