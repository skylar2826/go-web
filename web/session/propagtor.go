package session

import (
	"net/http"
)

type WebPropagator struct {
	cookieName string
}

func (w2 *WebPropagator) Inject(id string, w http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:  w2.cookieName,
		Value: id,
	}
	http.SetCookie(w, cookie)
	return nil
}

func (w2 *WebPropagator) Extract(r *http.Request) (string, error) {
	cookie, err := r.Cookie(w2.cookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (w2 *WebPropagator) Delete(w http.ResponseWriter) error {
	c := &http.Cookie{
		Name:   w2.cookieName,
		MaxAge: -1,
	}
	http.SetCookie(w, c)
	return nil
}

func NewWebPropagator(cookieName string) Propagator {
	return &WebPropagator{
		cookieName: cookieName,
	}
}
