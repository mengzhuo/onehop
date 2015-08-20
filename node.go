// Package onehop provides ...
package onehop

import (
	"fmt"
	"net"
	"time"
)

type Node struct {
	ID       string
	Addr     *net.UDPAddr
	updateAt time.Time
}

func (n *Node) String() string {
	return fmt.Sprintf("ID:%s Addr:%s", n.ID, n.Addr)
}
