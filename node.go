// Package onehop provides ...
package onehop

import (
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
	Type byte
	IP   [4]byte
	Port uint16
}

func (self *RemoteNode) RemoteToNode() *Node {

	id := new(big.Int)
	id.SetBytes(self.ID[:])
	ip := net.IPv4(self.IP[0], self.IP[1], self.IP[2], self.IP[3])
	return &Node{id, self.Type, ip, self.Port}

}

type Node struct {
	ID   *big.Int
	Type byte
	IP   net.IP
	Port uint16
}
