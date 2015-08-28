package onehop

import (
	"fmt"
	"net"
)

const (
	MSG_MAX_SIZE      = 8192
	MAX_EVENT_PER_MSG = MSG_MAX_SIZE / 310
	HEADER_LENGTH     = 4
)

// Message Type
const (
	_ = iota
	BOOTSTRAP
	BOOTSTRAP_RESPONSE
	KEEP_ALIVE
	KEEP_ALIVE_RESPONSE
	EVENT_NOTIFICATION
	MESSAGE_EXCHANGE
	REPLICATE
	REPLICATE_RESPONSE
	STORE
	FIND_VALUE
	FIND_VALUE_RESPONSE
)

var typeName = []string{
	"UNKNOWN",
	"BOOTSTRAP",
	"BOOTSTRAP_RESPONSE",
	"KEEP_ALIVE",
	"KEEP_ALIVE_RESPONSE",
	"EVENT_NOTIFICATION",
	"MESSAGE_EXCHANGE",
	"REPLICATE",
	"REPLICATE_RESPONSE",
	"STORE",
	"FIND_VALUE",
	"FIND_VALUE_RESPONSE",
}

// Event
const (
	_ byte = iota
	JOIN
	LEAVE
)

type Msg struct {
	ID     uint8    `json:"i"`
	Type   uint8    `json:"t"`
	From   string   `json:"f"`
	Time   int64    `json:"s"`
	Events []*Event `json:"e"`
}

func (m *Msg) String() string {

	return fmt.Sprintf("MSG F:%s T:%s E:%s", m.From, typeName[m.Type], m.Events)

}

type Event struct {
	ID     string       `json:"i"`
	Status uint8        `json:"s"`
	Addr   *net.UDPAddr `json:"a"` // IPv4 port
}

func (e *Event) String() string {
	return fmt.Sprintf("Event:S:%x A:%s", e.Status, e.Addr)
}
