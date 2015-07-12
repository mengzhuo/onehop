// Package onehop provides ...
package onehop

import (
	"fmt"
	"math/big"
	"net"
	"time"
)

// Node Type
const (
	_ byte = iota
	NORMAL
	UNIT_LEADER
	SLICE_LEADER
)

type Event struct {
	ID     *big.Int  `json:"i"`
	Time   time.Time `json:"t"`
	Status uint8     `json:"s"`
	Addr   string    `json:"a"`
}

func (self *Event) ToNode() *Node {

	addr, err := net.ResolveUDPAddr("udp", self.Addr)
	if err != nil {
		return nil
	}
	return &Node{ID: self.ID, Addr: addr}

}

type Node struct {
	ID       *big.Int
	Addr     *net.UDPAddr
	updateAt time.Time
	ticker   *time.Ticker
}

func (n *Node) resetTimer() {

	if n.ticker == nil {
		n.ticker = time.NewTicker(2 * time.Second)
		go n.startTick()
	}
	n.updateAt = time.Now()
}

func (n *Node) startTick() {

	for t := range n.ticker.C {
		if t.After(n.updateAt.Add(NODE_TIMEOUT)) {
			route.timeoutNode <- n
			break
		}
	}
}

func (n *Node) String() string {
	return fmt.Sprintf("ID:%016X Addr:%s", n.ID, n.Addr)
}

type ByID []*Node

func (b ByID) Len() int           { return len(b) }
func (b ByID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByID) Less(i, j int) bool { return b[i].ID.Cmp(b[j].ID) < 0 }
