// Package onehop provides ...
package onehop

import (
	"crypto/rand"
	"log"
	"math/big"
	"net"
	"time"
)

const (
	WRITE_TIME_OUT = 1 * time.Second
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

	id  *big.Int
	vid *big.Int // Vertical ID accross ring
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

	max := new(big.Int).SetBytes(FullID)
	half := new(big.Int).Div(max, big.NewInt(int64(2)))

	id, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}

	vid := new(big.Int).Add(id, half)
	if vid.Cmp(max) > 0 {
		vid.Mod(vid, max)
	}

	return &Service{listener, route, messagePool, bp, id, vid}
}

func (s *Service) Listen() {

	for {
		p := s.bytePool.Get()
		n, addr, err := s.conn.ReadFromUDP(p)

		log.Printf("From:%s Data:%x\n", addr, p[:n])

		if err != nil || n < 7 || p[0] != IDENTIFIER {
			s.bytePool.Put(p)
			continue
		}

		header := s.messagePool.Get()
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Error on get message %s", r)
				}
				// recycle
				s.messagePool.Put(header)
				s.bytePool.Put(p)
			}()

			header.Version = p[1]
			header.Type = p[2]
			header.ID = uint32(p[6]) | uint32(p[5])<<8 | uint32(p[4])<<16 | uint32(p[3])<<24
			log.Print("Header:", header)
			s.Handle(addr, header, p[6:])
		}()
	}
}

func readUint16(p []byte) uint16 {
	return uint16(p[1]) | uint16(p[0])<<8
}

func readID(p []byte) *big.Int {
	b := new(big.Int)
	b.SetBytes(p[:16])
	return b
}

func (s *Service) Send(dstAddr *net.UDPAddr, p []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Print(r)
		}
	}()

	buf := s.bytePool.Get()
	defer s.bytePool.Put(buf)

	buf[0] = IDENTIFIER
	buf = append(buf[:1], p...)
	s.conn.WriteToUDP(buf, dstAddr)
}
