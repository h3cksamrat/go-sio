gnabgib socket.io client implementation (in go)
================

golang implementation of [socket.io](http://socket.io) client

### Installation

    go get github.com/gnabgib/go-sio

### Client

```go
	//connect to server, you can use your own transport settings
	params:=make(map[string]string)
	t := transport.GetDefaultWebsocketTransport()
	ws, err := gosio.Dial(
		gosio.GetUrl("localhost", 3000, false, &params),
		t,
	)

	ws.Emit("hello", args)
	//do something, handlers and functions are same as server ones

	//close connection
	ws.Close()
```

### Dependencies

- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [glog](https://github.com/golang/glog) (shared with gorilla)
