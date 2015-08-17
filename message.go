package onehop

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

type MsgHeader struct {
	ID    uint8
	Type  uint8
	Count uint16
}

func ParseHeader(p []byte) *MsgHeader {

	h := new(MsgHeader)
	h.ID = p[0]
	h.Type = p[1]
	h.Count = uint16(p[2])<<8 | uint16(p[3])
	return h
}
