package gosio

import (
	"sync"

	"github.com/gnabgib/go-sio/protocol"
	"github.com/gnabgib/go-sio/transport"
)

// Channel - a collection of connection details
// - use Dial to connect to websocket
// - use IsAlive to check that handler is still working
// - use In and Out channels for message exchange
// - Close message means channel is closed
// - ping/pong replies are automatic
type Channel struct {
	conn transport.Connection

	in     chan *protocol.Message
	out    chan string
	header Header

	alive     bool
	aliveLock sync.Mutex

	ack ackProcessor
	sequentialInLoop bool
}

/**
create channel, map, and set active
*/
func (c *Channel) initChannel() {
	//TODO: queueBufferSize from constant to server or client variable
	c.in = make(chan *protocol.Message, queueBufferSize)
	c.out = make(chan string, queueBufferSize)
	c.ack.resultWaiters = make(map[int](chan string))
	c.alive = true
}

// ID - Of current connection (provided by server, unique)
func (c *Channel) ID() string {
	return c.header.Sid
}

// IsAlive - whether a channel is still alive
func (c *Channel) IsAlive() bool {
	c.aliveLock.Lock()
	defer c.aliveLock.Unlock()

	return c.alive
}
