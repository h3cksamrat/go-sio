package transport

import (
	"errors"
	"io/ioutil"
	"time"

	"github.com/gorilla/websocket"
)

var (
	errBinaryMessage = errors.New("Binary messages are not currently supported")
	errBadBuffer     = errors.New("Buffer error")
	errEmptyMessage  = errors.New("Empty message received")
)

// WebsocketConnection - A websocket connection
type WebsocketConnection struct {
	socket    *websocket.Conn
	transport *WebsocketTransport
	url string
}

func (ws *WebsocketConnection) String() string {
	return ws.url
}

// GetMessage - Receive a message (blocking)
func (ws *WebsocketConnection) GetMessage() (message string, err error) {
	ws.socket.SetReadDeadline(time.Now().Add(ws.transport.ReceiveTimeout))
	msgType, reader, err := ws.socket.NextReader()
	if err != nil {
		return "", err
	}

	//support only text messages exchange
	if msgType != websocket.TextMessage {
		return "", errBinaryMessage
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", errBadBuffer
	}
	text := string(data)

	//empty messages are not allowed
	if len(text) == 0 {
		return "", errEmptyMessage
	}

	return text, nil
}

// WriteMessage - Send a message (blocking)
func (ws *WebsocketConnection) WriteMessage(message string) error {
	ws.socket.SetWriteDeadline(time.Now().Add(ws.transport.SendTimeout))
	writer, err := ws.socket.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	if _, err := writer.Write([]byte(message)); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}

// Close connection
func (ws *WebsocketConnection) Close() {
	ws.socket.Close()
}

// PingParams - time interval and timeout settings for ping
func (ws *WebsocketConnection) PingParams() (interval, timeout time.Duration) {
	return ws.transport.PingInterval, ws.transport.PingTimeout
}
