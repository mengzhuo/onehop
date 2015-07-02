// Package onehop provides ...
package onehop

import (
	"net"
)

// Node Type
const (
	_ byte = iota
	NORMAL
	UNIT_LEADER
	SLICE_LEADER
)

type Node struct {
	ID   RecordID
	Type byte
	IP   [4]byte
	Port uint16
}

func (n *Node) GetIP() net.IP {
	return net.IPv4(n.IP[0], n.IP[1], n.IP[2], n.IP[3])
}
