package transport

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const (
	upgradeFailed         = "Upgrade failed: "
	defaultPingInterval   = 30 * time.Second
	defaultPingTimeout    = 60 * time.Second
	defaultReceiveTimeout = 60 * time.Second
	defaultSendTimeout    = 60 * time.Second
	defaultBufferSize     = 1024 * 32
)

var (
	errMethodNotAllowed  = errors.New("Method not allowed")
	errHTTPUpgradeFailed = errors.New("Http upgrade failed")
)

// WebsocketTransport - Connection factory for websocket
type WebsocketTransport struct {
	PingInterval   time.Duration
	PingTimeout    time.Duration
	ReceiveTimeout time.Duration
	SendTimeout    time.Duration

	BufferSize int

	RequestHeader http.Header
}

// Connect - Establish a new connection
func (wst *WebsocketTransport) Connect(url *url.URL) (conn Connection, err error) {
	dialer := websocket.DefaultDialer
	socket, _, err := dialer.Dial(url.String(), wst.RequestHeader)
	if err != nil {
		return nil, err
	}

	return &WebsocketConnection{socket, wst, url.Host}, nil
}

// HandleConnection -
func (wst *WebsocketTransport) HandleConnection(
	w http.ResponseWriter, r *http.Request) (conn Connection, err error) {

	if r.Method != http.MethodGet {
		http.Error(w, upgradeFailed+errMethodNotAllowed.Error(), http.StatusServiceUnavailable)
		return nil, errMethodNotAllowed
	}

	socket, err := websocket.Upgrade(w, r, nil, wst.BufferSize, wst.BufferSize)
	if err != nil {
		http.Error(w, upgradeFailed+err.Error(), 503)
		return nil, errHTTPUpgradeFailed
	}

	return &WebsocketConnection{socket, wst, ""}, nil
}

// Serve - noop (no further processing required for WS)
func (wst *WebsocketTransport) Serve(w http.ResponseWriter, r *http.Request) {}

// GetDefaultWebsocketTransport - Returns websocket connection with default interval/timeout settings
func GetDefaultWebsocketTransport() *WebsocketTransport {
	return &WebsocketTransport{
		PingInterval:   defaultPingInterval,
		PingTimeout:    defaultPingTimeout,
		ReceiveTimeout: defaultReceiveTimeout,
		SendTimeout:    defaultSendTimeout,
		BufferSize:     defaultBufferSize,
	}
}
