package realtime

import (
	"sync"
)

type ISessionStore interface {
	Add(uid string, client *Client)
	Remove(uid string) bool
	Get(uid string) *Client
	ForEach(f func(conn *Client))
}

type InMemorySessionStore struct {
	connections map[string]*Client
	mu          sync.RWMutex
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		connections: make(map[string]*Client),
	}
}

func (s *InMemorySessionStore) Add(uid string, client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connections[uid] = client
}

func (s *InMemorySessionStore) Remove(uid string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.connections[uid]
	if ok {
		delete(s.connections, uid)
		return true
	}

	return false
}

func (s *InMemorySessionStore) Get(uid string) *Client {
	client, ok := s.connections[uid]
	if ok {
		return client
	}

	return nil
}

func (s *InMemorySessionStore) ForEach(f func(conn *Client)) {
	for key := range s.connections {
		f(s.connections[key])
	}
}
