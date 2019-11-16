// +build !linux

package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type epoll struct {
	fd          int
	connections map[int]*websocket.Conn
	lock        *sync.RWMutex
}

func MkEpoll() (*epoll, error) {
	return nil, nil
}

func (e *epoll) Add(conn *websocket.Conn) error {
	return nil
}

func (e *epoll) Remove(conn *websocket.Conn) error {
	return nil
}

func (e *epoll) Wait() ([]*websocket.Conn, error) {
	return nil, nil
}
