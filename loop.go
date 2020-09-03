package gosio

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gnabgib/go-sio/protocol"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

const (
	queueBufferSize = 500
)

var (
	errorWrongHeader = errors.New("Wrong header")
)

// Header - engine.io header for messages
type Header struct {
	Sid          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
	PingInterval int      `json:"pingInterval"`
	PingTimeout  int      `json:"pingTimeout"`
}

func closeChannel(c *Channel, e *event, args ...interface{}) error {
	c.aliveLock.Lock()
	defer c.aliveLock.Unlock()

	if !c.alive {
		//already closed
		return nil
	}
	c.alive = false
	c.conn.Close()

	// close message in-channel
	close(c.in)

	//clean outloop
	for len(c.out) > 0 {
		<-c.out
	}
	c.out <- protocol.CloseMessage

	e.callLoopEvent(c, OnDisconnection)

	overfloodedLock.Lock()
	delete(overflooded, c)
	overfloodedLock.Unlock()

	return nil
}

//incoming messages loop, puts incoming messages to In channel
func inLoop(c *Channel, e *event) error {
	glog.V(4).Infoln("Start in loop for channel", c.conn)
	defer func() {
		glog.V(4).Infoln("Exit in loop for channel", c.conn)
	}()
	for {
		pkt, err := c.conn.GetMessage()

		if err != nil {
			if !websocket.IsCloseError(err,websocket.CloseNormalClosure,websocket.CloseGoingAway,websocket.CloseNoStatusReceived) {
				return nil
			}
			glog.Errorf("Failed to get message: %s", err)
			return closeChannel(c, e, err)
		}
		msg, err := protocol.Decode(pkt)
		if err != nil {
			glog.Errorf("Failed to decode message: %s", err)
			closeChannel(c, e, errors.New("Wrong packet"))
			return err
		}

		switch msg.Type {
		case protocol.MessageTypeOpen:
			if err := json.Unmarshal([]byte(msg.Source[1:]), &c.header); err != nil {
				glog.Errorf("Failed to decode message source: %s", err)
				closeChannel(c, e, errorWrongHeader)
			}
			e.callLoopEvent(c, OnConnection)
		case protocol.MessageTypePing:
			c.out <- protocol.PongMessage
		case protocol.MessageTypePong:
		default:
			glog.V(5).Infof("Received message %d %q", msg.Type, msg.Method)
			if c.sequentialInLoop {
				//glog.V(5).Infof("Process %q sequentially", msg.Method)
				c.in <- msg
			} else {
				//glog.V(5).Infof("Process %q asynchronously", msg.Method)
				go e.processIncomingMessage(c, msg)
			}
		}
	}
	return nil
}

// worker for processing messages
func workerLoop(c *Channel, e *event) error {
	glog.V(4).Infoln("Start worker loop for channel", c.conn)
	defer func() {
		glog.V(4).Infoln("Exit worker loop for channel", c.conn)
	}()
	for {
		select {
		case msg := <-c.in:
			if msg == nil {
				return nil
			}
			e.processIncomingMessage(c, msg)
		}
	}
}

var overflooded map[*Channel]struct{} = make(map[*Channel]struct{})
var overfloodedLock sync.Mutex

// func AmountOfOverflooded() int64 {
// 	overfloodedLock.Lock()
// 	defer overfloodedLock.Unlock()

// 	return int64(len(overflooded))
// }

/**
outgoing messages loop, sends messages from channel to socket
*/
func outLoop(c *Channel, e *event) error {
	glog.V(4).Infoln("Start out loop for channel", c.conn)
	defer func() {
		glog.V(4).Infoln("Exit out loop for channel", c.conn)
	}()
	for {
		outBufferLen := len(c.out)
		if outBufferLen >= queueBufferSize-1 {
			glog.Errorf("Output buffer to small")
			return closeChannel(c, e, errors.New("Buffer overlow"))
		} else if outBufferLen > int(queueBufferSize/2) {
			overfloodedLock.Lock()
			overflooded[c] = struct{}{}
			overfloodedLock.Unlock()
		} else {
			overfloodedLock.Lock()
			delete(overflooded, c)
			overfloodedLock.Unlock()
		}

		msg := <-c.out
		if msg == protocol.CloseMessage {
			return nil
		}

		err := c.conn.WriteMessage(msg)
		if err != nil {
			glog.Errorf("Failed to write message: %s", err)
			return closeChannel(c, e, err)
		}
	}
	return nil
}

/**
Pinger sends ping messages for keeping connection alive
*/
func pinger(c *Channel) {
	for {
		interval, _ := c.conn.PingParams()
		time.Sleep(interval)
		if !c.IsAlive() {
			return
		}

		c.out <- protocol.PingMessage
	}
}
