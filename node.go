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

const (
	ID_SIZE int = 16
)

type RecordID [ID_SIZE]byte

type RemoteNode struct {
	ID     RecordID
	Status byte
	IP     [4]byte
	Port   uint16
}

func (self *RemoteNode) ProLen() int {
	return ID_SIZE + 4 + 2
}

func (self *RemoteNode) ToNode() *Node {

	id := new(big.Int)
	id.SetBytes(self.ID[:])
	ip := net.IPv4(self.IP[0], self.IP[1], self.IP[2], self.IP[3])
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, self.Port))
	if err != nil {
		return nil
	}
	return &Node{ID: id, Addr: addr}

}

func NewRemoteNodeFromByte(p []byte) *RemoteNode {

	var id RecordID
	off := 0

	for i := 0; i < ID_SIZE; i++ {
		id[i] = p[i]
		off++
	}
	status := p[off]
	off++

	var ip [4]byte
	for i := 0; i < 4; i++ {
		ip[i] = p[off]
		off++
	}
	fmt.Printf("IP:%x", ip)
	port := uint16(p[off])<<8 | uint16(p[off+1])

	return &RemoteNode{id, status, ip, port}
}

type Node struct {
	ID     *big.Int
	Addr   *net.UDPAddr
	Ticker *time.Ticker
}

func (n *Node) String() string {
	return fmt.Sprintf("ID:%016X Addr:%s", n.ID, n.Addr)
}

type ByID []*Node

func (b ByID) Len() int           { return len(b) }
func (b ByID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByID) Less(i, j int) bool { return b[i].ID.Cmp(b[j].ID) < 0 }
