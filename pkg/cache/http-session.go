package cache

import (
	"context"
	"fmt"
	"net/http"
)

func NewSessionRepo[T any](
	cacheClient Client,
	cookieName string,
	sessionKeyPrefix string,
) func(http.ResponseWriter, *http.Request) *http.Request {
	repo := NewRepo[T](cacheClient)
	return func(w http.ResponseWriter, r *http.Request) *http.Request {
		cookie, _ := r.Cookie(cookieName)
		// TODO handle error
		key := fmt.Sprintf("%v:%v", sessionKeyPrefix, cookie)
		get, _ := repo.Get(r.Context(), key)
		// TODO handle error
		newContext := context.WithValue(r.Context(), "session", get)
		return r.WithContext(newContext)
	}
}
