package handler

import (
	"awesomeProject2/web/context"
	"fmt"
	"log"
	"sync"
)

type BasedOnMap struct {
	Handler sync.Map
}

func (h *BasedOnMap) ServeHTTP(c *context.Context) {
	key := h.Key(c.R.Method, c.R.URL.Path)

	if handlerFunc, ok := h.Handler.Load(key); !ok {
		err := c.NotFoundJson(key)
		log.Fatal(err)
	} else {
		c := context.NewContext(c.W, c.R)
		handlerFunc.(HandleFunc)(c)
	}
}

func (h *BasedOnMap) Route(method string, pattern string, handlerFunc HandleFunc) {
	key := h.Key(method, pattern)
	h.Handler.Store(key, handlerFunc)
}

func (h *BasedOnMap) Key(method string, path string) string {
	return fmt.Sprintf("%s#%s", method, path)
}

func NewHandlerBasedOnMap() *BasedOnMap {
	return &BasedOnMap{Handler: sync.Map{}}
}

var _ Handler = &BasedOnMap{}
