//go:build !linux
// +build !linux

package main

import (
	"log"
	"reflect"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
)

type epoll struct {
	fd          int
	connections map[int]*websocket.Conn
	changes     []syscall.Kevent_t
	lock        *sync.RWMutex
}

func MkEpoll() (*epoll, error) {
	fd, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}
	_, err = syscall.Kevent(fd, []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}}, nil, nil)
	if err != nil {
		return nil, err
	}
	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]*websocket.Conn),
	}, nil
}

func (e *epoll) Add(conn *websocket.Conn) error {
	fd := websocketFD(conn)
	e.changes = append(e.changes,
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ,
		},
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE,
		},
	)
	e.lock.Lock()
	defer e.lock.Unlock()
	e.connections[fd] = conn
	if len(e.connections)%100 == 0 {
		log.Printf("Total number of connections: %v", len(e.connections))
	}
	return nil
}

func (e *epoll) Remove(conn *websocket.Conn) error {
	fd := websocketFD(conn)
	for i, change := range e.changes {
		if change.Ident == uint64(fd) {
			e.changes[i].Flags = syscall.EV_DELETE
		}
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.connections, fd)
	if len(e.connections)%100 == 0 {
		log.Printf("Total number of connections: %v", len(e.connections))
	}
	return nil
}

func (e *epoll) Wait() ([]*websocket.Conn, error) {
	events := make([]syscall.Kevent_t, 100)
	timeout := syscall.Timespec{
		Sec:  0,
		Nsec: 0,
	}
	n, err := syscall.Kevent(e.fd, e.changes, events, &timeout)
	if err != nil {
		if err == syscall.EINTR {
			// timeout occurred
			return nil, nil
		}
		return nil, err
	}
	e.lock.RLock()
	defer e.lock.RUnlock()
	var connections []*websocket.Conn
	for i := 0; i < n; i++ {
		conn := e.connections[int(events[i].Ident)]
		connections = append(connections, conn)
	}
	changes := []syscall.Kevent_t{}
	for _, change := range e.changes {
		if change.Flags == syscall.EV_ADD {
			changes = append(changes, change)
		}
	}
	if len(e.changes) != len(changes) {
		e.changes = changes
	}
	return connections, nil
}

func websocketFD(conn *websocket.Conn) int {
	connVal := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn").Elem()
	tcpConn := reflect.Indirect(connVal).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
