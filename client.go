package gosio

import (
	"net/url"
	"strconv"

	"github.com/gnabgib/go-sio/transport"
)

const (
	webSocketScheme = "ws"
	pollingScheme   = "http"
	sioPath         = "/socket.io/"
)

// Client holds connection details
type Client struct {
	event
	Channel
}

// GetURL - Convert a host/port/secure flag and params into a URL
func GetURL(host string, port int, secure bool, params *map[string]string) *url.URL {
	url := url.URL{}
	url.Scheme = webSocketScheme
	url.Host = host + ":" + strconv.Itoa(port)
	url.Path = sioPath
	if secure {
		url.Scheme += "s"
	}

	q := url.Query()
	q.Set("EIO", "3")
	q.Set("transport", "websocket")
	for k, v := range *params {
		q.Set(k, v)
	}
	url.RawQuery = q.Encode()
	return &url
}

// Dial - connect to server and initialize protocol
// - You should use GetURL to generate the correct URL
func Dial(url *url.URL, tr transport.Transport) (*Client, error) {
	c := &Client{}
	c.initChannel()
	c.initMethods()

	var err error
	c.conn, err = tr.Connect(url)
	if err != nil {
		return nil, err
	}

	go workerLoop(&c.Channel, &c.event)
	go inLoop(&c.Channel, &c.event)
	go outLoop(&c.Channel, &c.event)
	go pinger(&c.Channel)

	return c, nil
}

// Dial2 - Similar to Dial, but set sequentialInLoop to true in Channel
// this will cause incoming message handling to be serialized.
func Dial2(url *url.URL, tr transport.Transport) (*Client, error) {
	c, err := Dial(url, tr)
	if err == nil {
		c.Channel.sequentialInLoop = true
	}
	return c, err
}

// Close client connection
func (c *Client) Close() {
	closeChannel(&c.Channel, &c.event)
}
