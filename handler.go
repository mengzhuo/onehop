package onehop

import (
	"log"
	"net"
)

func (s *Service) Handle(addr *net.UDPAddr, hdr *MessageHeader, body []byte) {
	// Main handle function for ALL msg
	switch hdr.Type {

	case KEEP_ALIVE:
		s.KeepAlive(addr, hdr, body)
	case EVENT_NOTIFICATION:

	case MESSAGE_EXCHANGE:

	case STORE:

	case FIND_NODE:

	case FIND_VALUE:

	case REPLICATE:

	default:
		log.Printf("UnKnown message %s %x", hdr, body)
	}
}

func (s *Service) KeepAlive(addr *net.UDPAddr, hdr *MessageHeader, p []byte) {
	count := int8(p[0])
	if count == 0 || count > 120 {
		return
	}
	rset := make([]RemoteNode, count)
	for i := int8(0); i < count; i++ {
		off := 22*int(i) + 1
		r := rset[i]
		for j := 0; j < len(r.ID); j++ {
			r.ID[j] = p[off+j]
		}
	}
}
