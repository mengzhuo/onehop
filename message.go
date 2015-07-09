package onehop

import "log"

const (
	IDENTIFIER   = 0x6f
	MSG_MAX_SIZE = 8192
	VERSION      = 0x01
)

// Message Type
const (
	_ = iota // 0x0 skipped
	BOOTSTRAP
	KEEP_ALIVE
	KEEP_ALIVE_RESPONSE
	EVENT_NOTIFICATION
	MESSAGE_EXCHANGE
	STORE
	FIND_VALUE
	FIND_NODE
	REPLICATE
)

// Event
const (
	_ byte = iota
	JOIN
	LEAVE
)

type MessageHeader struct {
	Version byte
	Type    byte
	ID      uint32
}

func (m *MessageHeader) ToBytes() []byte {
	b := make([]byte, 6)
	b[0] = m.Version
	b[1] = m.Type
	b[2] = byte(m.ID >> 24)
	b[3] = byte(m.ID >> 16)
	b[4] = byte(m.ID >> 8)
	b[5] = byte(m.ID)
	return b
}

func loadRecords(p []byte, off int) (records []RemoteNode, off1 int, ok bool) {

	if len(p[off:]) < 2 {
		log.Printf("Not enough msg")
		return records, off, false
	}

	off1 = off
	ok = false
	count := p[off1]
	off1++

	if count == 0 {
		return
	}
	for i := 0; i < int(count); i++ {
		record := NewRemoteNodeFromByte(p[off1 : off1+23])
		records = append(records, *record)
		off1 += 23
	}
	ok = true
	return
}

type KeepAliveMsg struct {
	Hdr         *MessageHeader
	From        RecordID
	RecordCount uint8
	Records     []RemoteNode
}

type EventNotificationMsg struct {
	Hdr         *MessageHeader
	RecordCount uint8
	Records     []RemoteNode
}

type ExchangeMsg struct {
	Hdr         *MessageHeader
	RecordCount uint8
	Records     []RemoteNode
}
