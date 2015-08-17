// Package onehop provides ...
package onehop

import (
	"fmt"
	"net"
	"time"
)

const EVENT_LENGTH = 27

type Event struct {
	ID     [16]byte
	Status uint8
	Time   int64
	Addr   [4]byte // IPv4 Addr
	Port   uint16  // IPv4 port
}

func ParseEvent(p []byte) (e *Event) {
	e = new(Event)
	copy(e.ID[:], p[0:16])
	e.Status = p[16]
	e.Time = int64(uint64(p[17])<<24 | uint64(p[18])<<16 | uint64(p[19])<<8 | uint64(p[20]))
	copy(e.Addr[:], p[21:25])
	e.Port = uint16(p[25])<<8 | uint16(p[26])
	return
}

func (e *Event) String() string {
	return fmt.Sprintf("Event:%s S:%x A:%s T:%s",
		e.ID, e.Status, e.Addr, e.Time)
}

type Node struct {
	ID       string
	Addr     *net.UDPAddr
	updateAt time.Time
}

func (n *Node) String() string {
	return fmt.Sprintf("ID:%s Addr:%s", n.ID, n.Addr)
}
