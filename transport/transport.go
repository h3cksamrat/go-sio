package transport

import (
	"net/http"
	"net/url"
	"time"
)

//Connection for given transport
type Connection interface {
	// GetMessage - Receive a message (blocking)
	GetMessage() (message string, err error)

	// WriteMessage - Send a message (blocking)
	WriteMessage(message string) error

	// Close connection
	Close()

	// PingParams - time interval and timeout settings for ping
	PingParams() (interval, timeout time.Duration)

	//Details of the connection
	String() string
}

//Transport - Connection factory for given transport
type Transport interface {
	// Connect - get client connection
	Connect(url *url.URL) (conn Connection, err error)

	// HandleConnection - Handle one server connection
	HandleConnection(w http.ResponseWriter, r *http.Request) (conn Connection, err error)

	//Serve HTTP request after establishing a connection
	Serve(w http.ResponseWriter, r *http.Request)
}
