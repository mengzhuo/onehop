package onehop

const (
	IDENTIFIER   = 0x31
	IDENTIFIER2  = 0x5e
	MSG_MAX_SIZE = 8192
)

// Message Type
const (
	_ = iota + 0x29 // "0"
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

type KeepAliveMsg struct {
	From        RecordID
	RecordCount int
	Records     []RemoteNode
}

type EventNotificationMsg struct {
	RecordCount int
	Records     []RemoteNode
}

type ExchangeMsg struct {
	RecordCount int
	Records     []RemoteNode
}
