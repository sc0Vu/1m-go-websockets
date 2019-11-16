package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/websocket"
)

var epoller *epoll

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	if err := epoller.Add(conn); err != nil {
		log.Printf("Failed to add connection")
		conn.Close()
	}
}

func main() {
	setNofileRlimit()
	// Enable pprof hooks
	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatalf("pprof failed: %v", err)
		}
	}()

	// Start epoll
	var err error
	epoller, err = MkEpoll()
	if err != nil {
		panic(err)
	}

	go Start()

	http.HandleFunc("/", wsHandler)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}

func Start() {
	for {
		connections, err := epoller.Wait()
		if err != nil {
			log.Printf("Failed to epoll wait %v", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if err := epoller.Remove(conn); err != nil {
					log.Printf("Failed to remove %v", err)
				}
				conn.Close()
			} else {
				log.Printf("msg: %s", string(msg))
			}
		}
	}
}
