package onehop

import (
	"fmt"
	"net"
	"time"
)

const (
	MSG_MAX_SIZE  = 8192
	HEADER_LENGTH = 4
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
	ID     uint8             `json:"i"`
	Type   uint8             `json:"t"`
	From   string            `json:"f"`
	Time   time.Time         `json:"s"`
	Events map[string]*Event `json:"e"`
}
type Event struct {
	Status uint8        `json:"s"`
	Addr   *net.UDPAddr `json:"a"` // IPv4 port
}

func (e *Event) String() string {
	return fmt.Sprintf("Event:S:%x A:%s", e.Status, e.Addr)
}
