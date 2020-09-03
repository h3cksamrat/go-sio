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
		gosio.GetURL("localhost", 3000, false, &params),
		t,
	)
	if err!=nil {
		fmt.Println(err)
		return
	}

	//Do something with the websocket
	ws.Emit("chat message","hi")
```

### Dependencies

- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [glog](https://github.com/golang/glog) (shared with gorilla)
