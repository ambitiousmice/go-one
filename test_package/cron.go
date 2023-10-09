package main

import (
	"flag"
	"go-one/common/idgenerator"
	"go-one/common/json"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  256,
	WriteBufferSize: 256,
	WriteBufferPool: &sync.Pool{},
}

func process(c *websocket.Conn) {
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	// Process connection in a new goroutine
	go process(c)

	// Let the http handler return, the 8k buffer created by it will be garbage collected
}

func main() {
	var GateInfo http.Request
	err := json.UnmarshalFromString("", GateInfo)
	if err == nil {

	}
	node, err := idgenerator.NewNode(1)

	println(node.NextIDStr())

}

type T struct {
	A int
}

var Map = make(map[int]*T)
