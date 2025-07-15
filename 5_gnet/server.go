package main

import (
	"github.com/cclehui/server_on_gnet/websocket"
	"github.com/panjf2000/gnet/v2"
	"log"
)

func main() {
	addr := "tcp://localhost:8000"
	wsServer := websocket.NewEchoServer(addr)
	err := gnet.Run(wsServer, addr, gnet.WithMulticore(true))
	if err != nil {
		log.Fatal(err)
	}
}
