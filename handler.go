package onehop

import (
	"log"
	"net"
)

func (s *Service) Handle(raddr *net.UDPAddr, hdr *MessageHeader, body []byte) {

	// Main handle function for ALL msg
	switch hdr.Type {

	case KEEP_ALIVE:
		s.KeepAlive(raddr, hdr, body)
	case KEEP_ALIVE_RESPONSE:
		s.KeepAliveResponse(raddr, hdr, body)
	default:
		log.Printf("UnKnown message %s %x", hdr, body)
	}
}

func (s *Service) KeepAliveResponse(raddr *net.UDPAddr, hdr *MessageHeader, p []byte) {

}

func (s *Service) KeepAlive(raddr *net.UDPAddr, hdr *MessageHeader, p []byte) {

	from := readID(p)
	records, _, ok := loadRecords(p, 16)

	if !ok {
		return
	}

	for _, rn := range records {

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

	i := unit.getID(from)
	cmp := from.Cmp(s.id)

	switch {
	case cmp == -1 && i > 0:
		i--
	case cmp == 1 && (i < unit.Len()-1):
		i++
	}
	// Send to successor/predecessor in unit
	n := unit.nodes[i]
	s.Send(n.Addr, []byte{})
	// Response to requester
	s.Send(raddr, hdr.ToBytes())
}
