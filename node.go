// Package onehop provides ...
package onehop

import (
	"fmt"
	"net"
	"sync"
)

type Node struct {
	ID       string
	Addr     *net.UDPAddr
	updateAt int64
	*sync.Mutex
}

func (n *Node) String() string {
	return fmt.Sprintf("ID:%s Addr:%s", n.ID, n.Addr)
}

func (n *Node) Update(u int64) {
	n.Lock()
	defer n.Unlock()
	n.updateAt = u
}
