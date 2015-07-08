package onehop

const IDENTIFIER = 0x6f

// Message Type
const (
	_ = iota // 0x0 skipped
	BOOTSTRAP
	KEEP_ALIVE
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
}

type KeepAliveMsg struct {
	Hdr         *MessageHeader
	RecordCount uint8
	Records     []RemoteNode
}

type EventNotificationMsg struct {
	Hdr         *MessageHeader
	RecordCount uint8
	Records     []RemoteNode
}

type MessageExchangeMsg struct {
	Hdr         *MessageHeader
	RecordCount uint8
	Records     []RemoteNode
}
