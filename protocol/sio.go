package protocol

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

const (
	open = "0"
	//CloseMessage - Request to close connection
	CloseMessage = "1"
	//PingMessage - Ping request
	PingMessage = "2"
	//PongMessage - Pong reply
	PongMessage = "3"
	msg         = "4"
	msgEmpty    = "0" //Append after msg (40)
	msgCommon   = "2" //Append after msg (42)
	msgAck      = "3" //Append after msg (43)

)

var (
	errorUnknownMessageType = errors.New("Unknown message type")
	errorWrongPacket        = errors.New("Wrong packet")
)

func typeToText(msgType int) (string, error) {
	switch msgType {
	case MessageTypeOpen:
		return open, nil
	case MessageTypeClose:
		return CloseMessage, nil
	case MessageTypePing:
		return PingMessage, nil
	case MessageTypePong:
		return PongMessage, nil
	case MessageTypeEmpty:
		return msg + msgEmpty, nil
	case MessageTypeEmit, MessageTypeAckRequest:
		return msg + msgCommon, nil
	case MessageTypeAckResponse:
		return msg + msgAck, nil
	}
	return "", errorUnknownMessageType
}

// Encode - Convert a message into a string for the wire
func Encode(m *Message) (string, error) {
	mtype, err := typeToText(m.Type)
	if err != nil {
		return "", err
	}

	switch m.Type {
	case MessageTypeEmpty, MessageTypePing, MessageTypePong:
		return mtype, nil
	case MessageTypeAckRequest:
		mtype += strconv.Itoa(m.AckID)
	case MessageTypeAckResponse:
		return mtype + strconv.Itoa(m.AckID) + "[" + m.Args + "]", nil
	case MessageTypeOpen, MessageTypeClose:
		return mtype + m.Args, nil
	}

	jsonMethod, err := json.Marshal(&m.Method)
	if err != nil {
		return "", err
	}

	return mtype + "[" + string(jsonMethod) + "," + m.Args + "]", nil
}

// MustEncode - Encode or panic
func MustEncode(m *Message) string {
	result, err := Encode(m)
	if err != nil {
		panic(err)
	}

	return result
}

func getMessageType(data string) (int, error) {
	if len(data) == 0 {
		return 0, errorUnknownMessageType
	}
	switch data[0:1] {
	case open:
		return MessageTypeOpen, nil
	case CloseMessage:
		return MessageTypeClose, nil
	case PingMessage:
		return MessageTypePing, nil
	case PongMessage:
		return MessageTypePong, nil
	case msg:
		if len(data) == 1 {
			return 0, errorUnknownMessageType
		}
		switch data[1:2] {
		case msgEmpty:
			return MessageTypeEmpty, nil
		case msgCommon:
			return MessageTypeAckRequest, nil
		case msgAck:
			return MessageTypeAckResponse, nil
		}
	}
	return 0, errorUnknownMessageType
}

/*
*
Get ack id of current packet, if present
*/
func getAck(text string) (AckID int, restText string, err error) {
	if len(text) < 4 {
		return 0, "", errorWrongPacket
	}
	text = text[2:]

	pos := strings.IndexByte(text, '[')
	if pos == -1 {
		return 0, "", errorWrongPacket
	}

	ack, err := strconv.Atoi(text[0:pos])
	if err != nil {
		return 0, "", err
	}

	return ack, text[pos:], nil
}

/*
*
Get message method of current packet, if present
*/
func getMethod(text string) (method, restText string, err error) {
	// var args []string

	// err = json.Unmarshal([]byte(text), &args)
	// if err != nil {
	// 	return "", "", err
	// }
	// if len(args) < 2 {
	// 	return "", "", errorWrongPacket
	// }

	// return args[0], args[1], err
	var start, end, rest, countQuote int

	text = strings.TrimSpace(text)
	for i, c := range text {
		if c == '"' {
			switch countQuote {
			case 0:
				start = i + 1
			case 1:
				end = i
				rest = i + 1
			default:
				return "", "", errorWrongPacket
			}
			countQuote++
		}
		if c == ',' {
			if countQuote < 2 {
				continue
			}
			rest = i + 1
			break
		}
	}

	if (end < start) || (rest >= len(text)) {
		return "", "", errorWrongPacket
	}

	return text[start:end], text[rest : len(text)-1], nil
}

// Decode - take a message from the wire and convert it back into the Message struct
func Decode(data string) (*Message, error) {
	var err error
	m := &Message{Source: data}

	m.Type, err = getMessageType(data)
	if err != nil {
		return nil, err
	}

	switch m.Type {
	case MessageTypeOpen:
		m.Args = data[1:]
		return m, nil
	case MessageTypeClose, MessageTypePing, MessageTypePong, MessageTypeEmpty:
		return m, nil
	}

	ack, rest, err := getAck(data)
	m.AckID = ack
	if m.Type == MessageTypeAckResponse {
		if err != nil {
			return nil, err
		}
		m.Args = rest[1 : len(rest)-1]
		return m, nil
	}

	if err != nil {
		m.Type = MessageTypeEmit
		rest = data[2:]
	}

	m.Method, m.Args, err = getMethod(rest)
	if err != nil {
		return nil, err
	}

	return m, nil
}
