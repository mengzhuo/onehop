// Package onehop provides ...
package onehop

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net"
	"time"
)

const (
	WRITE_TIME_OUT = 1 * time.Second
)

var (
	route   *Route
	selfIDs map[*big.Int]bool
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

	bytePool *BytePool

	id  *big.Int
	vid *big.Int // Vertical ID accross ring
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

	log.Printf("Listening to :%s %s", netType, address)
	route = NewRoute(k, u)
	messagePool := &MessageHeaderPool{make(chan *MessageHeader, 1024)}
	bp := &BytePool{make(chan []byte, 1024), 8192}

	max := new(big.Int).SetBytes(FullID)
	half := new(big.Int).Div(max, big.NewInt(int64(2)))

	id, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}

	n := &Node{ID: id, Addr: listener.LocalAddr().(*net.UDPAddr)}
	slice_idx, unit_idx := route.GetIndex(n.ID)
	slice := route.slices[slice_idx]
	unit := slice.units[unit_idx]
	if !unit.add(n) {
		fmt.Errorf("Can't not add n%s", n)
	}

	vid := new(big.Int).Add(id, half)
	if vid.Cmp(max) > 0 {
		vid.Mod(vid, max)
	}

	n = &Node{ID: vid, Addr: listener.LocalAddr().(*net.UDPAddr)}
	slice_idx, unit_idx = route.GetIndex(n.ID)
	slice = route.slices[slice_idx]
	unit = slice.units[unit_idx]
	if !unit.add(n) {
		fmt.Errorf("Can't not add n%s", n)
	}

	return &Service{listener, route, bp, id, vid}
}

func (s *Service) Listen() {

	for {
		p := s.bytePool.Get()
		n, addr, err := s.conn.ReadFromUDP(p)

		if err != nil || n < 7 || p[0] != IDENTIFIER || p[1] != IDENTIFIER2 {
			log.Printf("Not Enought for parsing...", addr, p[:n])
			s.bytePool.Put(p)
			continue
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Error on get message %s", r)
				}
				// recycle
				s.bytePool.Put(p)
			}()
			s.Handle(addr, p[2], p[2:6], p[6:n])
		}()
	}
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
