package gosio

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/golang/glog"
	"github.com/gnabgib/go-sio/protocol"
)

const (
	//OnConnection - Connect event
	OnConnection    = "connection"
	//OnDisconnection - Disconnect event
	OnDisconnection = "disconnection"
	//OnError - Error event
	OnError         = "error"
)

/**
Contains maps of message processing functions
*/
type event struct {
	messageHandlers     map[string]*caller
	messageHandlersLock sync.RWMutex

	onConnection    systemHandler
	onDisconnection systemHandler
}

/**
create messageHandlers map
*/
func (e *event) initMethods() {
	e.messageHandlers = make(map[string]*caller)
}

/**
Add message processing function, and bind it to given method
*/
func (e *event) On(method string, f interface{}) error {
	glog.V(5).Info("Listening to ",method)
	c, err := newCaller(f)
	if err != nil {
		return err
	}

	e.messageHandlersLock.Lock()
	e.messageHandlers[method] = c
	e.messageHandlersLock.Unlock()

	return nil
}
func (e *event) OnConnect(f systemHandler)  {
	e.messageHandlersLock.Lock()
	e.onConnection=f
	e.messageHandlersLock.Unlock()
}
func (e *event) OnDisconnect(f systemHandler) {
	e.messageHandlersLock.Lock()
	e.onDisconnection=f
	e.messageHandlersLock.Unlock()
}

/**
Find message processing function associated with given method
*/
func (e *event) findMethod(method string) (*caller, bool) {
	e.messageHandlersLock.RLock()
	defer e.messageHandlersLock.RUnlock()

	f, ok := e.messageHandlers[method]
	return f, ok
}

func (e *event) callLoopEvent(c *Channel, event string) {
	if event == OnConnection {
		if e.onConnection != nil {
			e.onConnection(c)
		}
		return
	}
	if event == OnDisconnection {
		if e.onDisconnection != nil {
			e.onDisconnection(c)
		}
		return
	}

	f, ok := e.findMethod(event)
	if !ok {
		return
	}

	f.callFunc(c, &struct{}{})
}

/**
Check incoming message
On ack_resp - look for waiter
On ack_req - look for processing function and send ack_resp
On emit - look for processing function
*/
func (e *event) processIncomingMessage(c *Channel, msg *protocol.Message) {
	switch msg.Type {
	case protocol.MessageTypeEmit:
		glog.V(5).Info("got-emit:",msg.Method,"(",msg.Args,")")
		f, ok := e.findMethod(msg.Method)
		if !ok {
			glog.V(5).Info("Couldn't find message for ",msg.Method)
			return
		}

		if !f.ArgsPresent {
			f.callFunc(c, &struct{}{})
			return
		}

		data := f.getArgs()
		
		err := json.Unmarshal([]byte(msg.Args), &data)
		if err != nil {
			glog.V(5).Info(msg.Method,"Unable to decode reply",err)
			return		
		}

		f.callFunc(c, data)

	case protocol.MessageTypeAckRequest:
		glog.V(5).Info("got-ack:",msg.Method,"(",msg.Args,")")
		f, ok := e.findMethod(msg.Method)
		if !ok || !f.Out {
			return
		}

		var result []reflect.Value
		if f.ArgsPresent {
			//data type should be defined for unmarshall
			data := f.getArgs()
			err := json.Unmarshal([]byte(msg.Args), &data)
			if err != nil {
				return
			}

			result = f.callFunc(c, data)
		} else {
			result = f.callFunc(c, &struct{}{})
		}

		ack := &protocol.Message{
			Type:  protocol.MessageTypeAckResponse,
			AckID: msg.AckID,
		}
		send(ack, c, result[0].Interface())

	case protocol.MessageTypeAckResponse:
		glog.V(5).Info("got-ack-response:",msg.AckID)
		waiter, err := c.ack.getWaiter(msg.AckID)
		if err == nil {
			waiter <- msg.Args
		}
	}
}
