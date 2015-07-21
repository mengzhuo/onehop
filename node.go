// Package onehop provides ...
package onehop

import (
	"fmt"
	"math/big"
	"net"
	"time"
)

type Event struct {
	ID     *big.Int  `json:"i"`
	Time   time.Time `json:"t"`
	Status uint8     `json:"s"`
	Addr   string    `json:"a"`
}

func (e *Event) String() string {
	return fmt.Sprintf("Event:%x S:%x A:%s T:%s",
		e.ID, e.Status, e.Addr, e.Time)
}

func (self *Event) ToNode() *Node {

	addr, err := net.ResolveUDPAddr("udp", self.Addr)
	if err != nil {
		return nil
	}
	return &Node{ID: self.ID, Addr: addr, updateAt: time.Now()}

}

type Node struct {
	ID       *big.Int
	Addr     *net.UDPAddr
	updateAt time.Time
}

func (n *Node) String() string {
	return fmt.Sprintf("ID:%016X Addr:%s", n.ID, n.Addr)
}
