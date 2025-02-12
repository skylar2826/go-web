package handler

import (
	"awesomeProject2/web/context"
)

type HandleFunc func(c *context.Context)

type Routable interface {
	Route(method string, pattern string, handlerFunc HandleFunc)
}

type Handler interface {
	ServeHTTP(c *context.Context)
	Routable
}
