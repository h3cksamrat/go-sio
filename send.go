package gosio

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gnabgib/go-sio/protocol"
	"github.com/golang/glog"
)

var (
	errorSendTimeout   = errors.New("Timeout")
	errorBufferOverlow = errors.New("Buffer overflow")
)

/**
Send message packet to socket
*/
func send(msg *protocol.Message, c *Channel, args interface{}) error {
	//preventing json/encoding "index out of range" panic
	defer func() {
		if r := recover(); r != nil {
			log.Println("socket.io send panic: ", r)
		}
	}()

	if args != nil {
		json, err := json.Marshal(&args)
		if err != nil {
			return err
		}

		msg.Args = string(json)
	}

	command, err := protocol.Encode(msg)
	if err != nil {
		return err
	}

	if len(c.out) == queueBufferSize {
		return errorBufferOverlow
	}

	glog.V(5).Info("Sending ",command)
	c.out <- command

	return nil
}

// Emit - Send a message to the server (do not expect a response)
func (c *Channel) Emit(method string, args interface{}) error {
	msg := &protocol.Message{
		Type:   protocol.MessageTypeEmit,
		Method: method,
	}

	return send(msg, c, args)
}

// Ack - Send a message to the server, expect a response
func (c *Channel) Ack(method string, args interface{}, timeout time.Duration) (string, error) {
	msg := &protocol.Message{
		Type:   protocol.MessageTypeAckRequest,
		AckID:  c.ack.nextID(),
		Method: method,
	}

	waiter := make(chan string)
	c.ack.addWaiter(msg.AckID, waiter)

	err := send(msg, c, args)
	if err != nil {
		c.ack.removeWaiter(msg.AckID)
	}

	select {
	case result := <-waiter:
		return result, nil
	case <-time.After(timeout):
		c.ack.removeWaiter(msg.AckID)
		return "", errorSendTimeout
	}
}
