package onehop

import (
	"net"
)

const (
	_ = iota // 0x0 skipped
	FIND_VALUE
	ALIVE
	LEAVE
)

type MessageHeader struct {
	Version byte
	Type    byte
}

type RecordID [16]byte

type Record struct {
	ID     RecordID
	Status byte
	IP     net.IP
	Port   uint16
}

type KeepAliveMessage struct {
	Hdr         MessageHeader
	RecordCount uint8
	Records     []Record
}
