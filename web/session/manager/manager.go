package manager

import (
	"awesomeProject2/web/context"
	"awesomeProject2/web/session"
	"awesomeProject2/web/session/session_store"
	"github.com/redis/go-redis/v9"
	"time"
)

type Manager struct {
	session.Store
	session.Propagator
}

func (m *Manager) InitSession(c *context.Context, id string) (session.Session, error) {
	s, err := m.Generate(c, id)
	if err != nil {
		return nil, err
	}
	err = m.Inject(id, c.W)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (m *Manager) GetSession(c *context.Context) (session.Session, error) {
	id, err := m.Extract(c.R)
	if err != nil {
		return nil, err
	}
	var sess session.Session
	if c.UserValues == nil {
		c.UserValues = make(map[string]any, 1)
	}
	if val, ok := c.UserValues[id]; ok {
		return val.(session.Session), nil
	}
	sess, err = m.Get(c, id)
	if err != nil {
		return nil, err
	}
	c.UserValues[id] = sess

	return sess, nil
}

func (m *Manager) RefreshSession(c *context.Context, id string) error {
	return m.Refresh(c, id)
}

func (m *Manager) RemoveSession(c *context.Context, id string) error {
	err := m.Remove(c, id)
	delete(c.UserValues, id)
	if err != nil {
		return err
	}
	return m.Delete(c.W)
}

func NewManager(cookieName string, store session.Store) *Manager {
	return &Manager{
		Store: store,
		//Store:      session_store.NewMemoryStore(time.Minute),
		Propagator: session.NewWebPropagator(cookieName),
	}
}

var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})
var store = session_store.NewRedisStore(rdb, "session_id", time.Minute*2)
var WebManager = NewManager("session_id", store)
