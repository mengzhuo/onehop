package onehop

// Message Type
const (
	_ = iota // 0x0 skipped
	BOOTSTRAP
	KEEP_ALIVE
	EVENT_NOTIFICATION
	MESSAGE_EXCHANGE
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

type KeepAliveMessage struct {
	Hdr         MessageHeader
	RecordCount uint8
	Records     []RemoteNode
}
