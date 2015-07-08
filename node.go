// Package onehop provides ...
package onehop

import (
	"fmt"
	"math/big"
	"net"
)

// Node Type
const (
	_ byte = iota
	NORMAL
	UNIT_LEADER
	SLICE_LEADER
)

const (
	ID_SIZE int = 16
)

type RecordID [ID_SIZE]byte

type RemoteNode struct {
	ID   RecordID
	IP   [4]byte
	Port uint16
}

func (self *RemoteNode) RemoteToNode() *Node {

	id := new(big.Int)
	id.SetBytes(self.ID[:])
	ip := net.IPv4(self.IP[0], self.IP[1], self.IP[2], self.IP[3])
	return &Node{id, ip, self.Port}

}

type Node struct {
	ID   *big.Int
	IP   net.IP
	Port uint16
}

func (n *Node) String() string {
	return fmt.Sprintf("ID:%016X Addr:%s:%d", n.ID, n.IP, n.Port)
}

type ByID []*Node

func (b ByID) Len() int           { return len(b) }
func (b ByID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByID) Less(i, j int) bool { return b[i].ID.Cmp(b[j].ID) < 0 }
