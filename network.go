// Package onehop provides ...
package onehop

import (
	"net"
)

func NewListener(netType, address string) (conn *net.UDPConn, err error) {

	uaddr, err := net.ResolveUDPAddr(netType, address)
	if err != nil {
		return nil, err
	}

	conn, err = net.ListenUDP(netType, uaddr)
	if err != nil {
		return nil, err
	}

	return conn, err
}

type Server struct {
	listener *net.UDPConn
	route    *Route

	messagePool chan *MessageHeader
}
