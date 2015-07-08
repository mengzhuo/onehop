// Package onehop provides ...
package onehop

import (
	"net"
)

func newListener(netType, address string) (conn *net.UDPConn, err error) {

	uaddr, err := net.ResolveUDPAddr(netType, address)
	if err != nil {
		return nil, err
	}

	conn, err = net.ListenUDP(netType, uaddr)
	if err != nil {
		return nil, err
	}

	return conn, err
}

type Service struct {
	conn  *net.UDPConn
	route *Route

	messagePool *MessageHeaderPool
	bytePool    *BytePool
}

type MessageHeaderPool struct {
	freeList chan *MessageHeader
}

func (p *MessageHeaderPool) Get() (h *MessageHeader) {
	select {
	case h = <-p.freeList:
	default:
		h = new(MessageHeader)
	}
	return
}

func (p *MessageHeaderPool) Put(h *MessageHeader) {
	select {
	case p.freeList <- h:
	default:
	}
}

type BytePool struct {
	freeList chan []byte
	maxSize  int
}

func (b *BytePool) Get() (p []byte) {
	select {
	case p = <-b.freeList:
	default:
		p = make([]byte, b.maxSize)

	}
	return
}

func (b *BytePool) Put(p []byte) {
	select {
	case b.freeList <- p:
	default:
	}
}

// NetType, Address for UDP connection
// k, u for OneHop slice number and unit number
func NewService(netType, address string, k, u int) *Service {

	listener, err := newListener(netType, address)
	if err != nil {
		panic(err)
	}

	route := NewRoute(k, u)
	messagePool := &MessageHeaderPool{make(chan *MessageHeader, 1024)}
	bp := &BytePool{make(chan []byte, 1024), 8192}
	return &Service{listener, route, messagePool, bp}
}

func (s *Service) Listen() {

	for {
		p := s.bytePool.Get()
		n, addr, err := s.conn.ReadFromUDP(p)

		if err != nil || n < 4 || p[0] != IDENTIFIER {
			s.bytePool.Put(p)
			continue
		}

		header := s.messagePool.Get()
		go func() {
			defer func() {
				// recycle
				s.messagePool.Put(header)
				s.bytePool.Put(p)
			}()

			header.Version = p[1]
			header.Type = p[2]
			s.Handle(addr, header, p[3:])
		}()
	}
}
