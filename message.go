package onehop

import (
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
	STORE
	FIND_VALUE
	FIND_VALUE_RESPONSE
)

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

func (m *Msg) NewID() {
	m.ID = rand.Uint32()
}

type MsgPool struct {
	freeList chan *Msg
}

func (p *MsgPool) Get() (m *Msg) {
	select {
	case m = <-p.freeList:
	default:
		m = new(Msg)
	}
	return m
}

func (p *MsgPool) Put(m *Msg) {

	// Clear all fields for not mess around
	m.Type = 0x0
	m.ID = 0
	m.Events = nil
	m.From = nil

	select {
	case p.freeList <- m:
	default:
	}
}
