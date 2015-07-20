package onehop

import (
	"fmt"
	"math/big"
	"math/rand"
)

const (
	IDENTIFIER   = 0x31
	MSG_MAX_SIZE = 8192
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
	ID     uint32   `json:"i"`
	Type   uint8    `json:"t"`
	From   *big.Int `json:"f"` // ID of remote node
	Events []Event  `json:"e,omitempty"`
}

func (m *Msg) String() string {
	return fmt.Sprintf("Msg:%d,Type:%s, From:%x, Events:%s",
		m.ID, typeName[m.Type], m.From, m.Events)

}

func NewMsg(typ uint8, f *big.Int, es []Event) *Msg {
	msg := new(Msg)
	msg.NewID()
	msg.Type = typ
	msg.From = f
	msg.Events = es
	return msg

}

func (m *Msg) NewID() {
	m.ID = rand.Uint32()
}
