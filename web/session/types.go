package session

import (
	"awesomeProject2/web/context"
	"net/http"
)

type Session interface {
	Get(c *context.Context, key string) (interface{}, error)
	Set(c *context.Context, key string, val interface{}) error
	ID() string
}

type Store interface {
	Generate(c *context.Context, id string) (Session, error)
	Get(c *context.Context, id string) (Session, error)
	Remove(c *context.Context, id string) error
	Refresh(c *context.Context, id string) error
}

type Propagator interface {
	Inject(id string, w http.ResponseWriter) error
	Extract(r *http.Request) (string, error)
	Delete(w http.ResponseWriter) error
}
