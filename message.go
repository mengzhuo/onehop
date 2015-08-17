package onehop

import (
	"fmt"
	"math/rand"
)

const MSG_MAX_SIZE = 8192

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
	ID     int
	Type   uint8
	From   string
	Events []Event
}

func (m *Msg) String() string {
	return fmt.Sprintf("Msg:%d,Type:%s, From:%x, Events:%s",
		m.ID, typeName[m.Type], m.From, m.Events)

}

func NewMsg(typ uint8, f string, es []Event) *Msg {
	msg := new(Msg)
	msg.NewID()
	msg.Type = typ
	msg.From = f
	msg.Events = es
	return msg

}

func (m *Msg) NewID() {
	m.ID = rand.Int()
}
