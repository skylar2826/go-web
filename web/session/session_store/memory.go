package session_store

import (
	"awesomeProject2/web/context"
	"awesomeProject2/web/custom_error"
	"awesomeProject2/web/session"
	"fmt"
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
)

type memorySession struct {
	id  string
	val sync.Map
}

func (m *memorySession) Get(c *context.Context, key string) (interface{}, error) {
	val, ok := m.val.Load(key)
	if !ok {
		return "", custom_error.ErrorNotFound(fmt.Sprintf("session key %s\n", key))
	}
	return val, nil
}

func (m *memorySession) Set(c *context.Context, key string, val interface{}) error {
	m.val.Store(key, val)
	return nil
}

func (m *memorySession) ID() string {
	return m.id
}

type MemoryStore struct {
	sessions *cache.Cache
	expired  time.Duration
	mutex    sync.Mutex
}

func (m *MemoryStore) Generate(c *context.Context, id string) (session.Session, error) {
	s := &memorySession{
		id: id,
	}
	m.mutex.Lock()
	m.sessions.Set(id, s, m.expired)
	m.mutex.Unlock()
	return s, nil
}

func (m *MemoryStore) Get(c *context.Context, id string) (session.Session, error) {
	m.mutex.Lock()
	s, ok := m.sessions.Get(id)
	m.mutex.Unlock()
	if !ok {
		return nil, custom_error.ErrorNotFound(fmt.Sprintf("session key %s\n", id))
	}
	return s.(session.Session), nil
}

func (m *MemoryStore) Remove(c *context.Context, id string) error {
	m.mutex.Lock()
	m.sessions.Delete(id)
	m.mutex.Unlock()
	return nil
}

func (m *MemoryStore) Refresh(c *context.Context, id string) error {
	s, ok := m.sessions.Get(id)
	if !ok {
		return custom_error.ErrorNotFound(fmt.Sprintf("session key %s\n", id))
	}
	m.mutex.Lock()
	m.sessions.Set(id, s, m.expired)
	m.mutex.Unlock()
	return nil
}

func NewMemoryStore(expired time.Duration) session.Store {
	return &MemoryStore{
		expired:  expired,
		sessions: cache.New(expired, time.Second*20),
	}
}
