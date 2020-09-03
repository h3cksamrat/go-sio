package protocol

const (
	// MessageTypeOpen - Connection opening
	MessageTypeOpen = iota
	// MessageTypeClose - Connection closing
	MessageTypeClose
	// MessageTypePing - Ping (see pong)
	MessageTypePing
	// MessageTypePong - Pong (see ping)
	MessageTypePong
	// MessageTypeEmpty - Empty
	MessageTypeEmpty
	// MessageTypeEmit - Emit message
	MessageTypeEmit 
	// MessageTypeAckRequest - Request ack message
	MessageTypeAckRequest
	// MessageTypeAckResponse - Reply to ack
	MessageTypeAckResponse
)

// Message - a message
type Message struct {
	Type   int
	AckID  int
	Method string
	Args   string
	Source string
}
