package stores

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Store[T any] interface {
	Add() T
	Remove()
	Get() *T
}

type InMemorySessionStore struct {
	connections map[string]*websocket.Conn
	mu          sync.RWMutex
}

func (s *InMemorySessionStore) Add(userid string, conn *websocket.Conn) {
	s.mu.Lock()
	s.connections[userid] = conn
	s.mu.Unlock()
}

func (s *InMemorySessionStore) Remove(userid string) bool {
	s.mu.Lock()

	_, ok := s.connections[userid]
	if !ok {
		return false
	}
	delete(s.connections, userid)
	s.mu.Unlock()

	return true
}

func (s *InMemorySessionStore) Get(userid string) *websocket.Conn {
	conn, ok := s.connections[userid]
	if !ok {
		return nil
	}

	return conn
}
