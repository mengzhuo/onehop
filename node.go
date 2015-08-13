// Package onehop provides ...
package onehop

import (
	"fmt"
	"net"
	"time"
)

type Event struct {
	ID     string    `json:"i"`
	Time   time.Time `json:"t"`
	Status uint8     `json:"s"`
	Addr   string    `json:"a"`
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
