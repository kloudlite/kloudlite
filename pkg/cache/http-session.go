package cache

import (
	"context"
	"fmt"
	"kloudlite.io/pkg/repos"
	"net/http"
	"time"
)

func NewSessionRepo[T repos.Entity](
	cacheClient Client,
	cookieName string,
	cookieDomain string,
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
		newContext = context.WithValue(newContext, "set-session", func(session T) {
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    string(session.GetId()),
				Path:     "/",
				Domain:   fmt.Sprintf(".%v", cookieDomain),
				Expires:  time.Time{},
				MaxAge:   0,
				Secure:   false,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
		})
		newContext = context.WithValue(newContext, "delete-session", func() {
			repo.Drop(newContext, key)
		})
		return r.WithContext(newContext)
	}
}

func GetSession[T repos.Entity](ctx context.Context) T {
	return ctx.Value("session").(T)
}

func SetSession[T repos.Entity](ctx context.Context, session T) {
	setSession, ok := ctx.Value("set-session").(func(T))
	fmt.Println("HERERERE", ok)
	if !ok {
		return
	}
	setSession(session)
}

func DeleteSession[T any](ctx context.Context) {
	deleteSession, ok := ctx.Value("delete-session").(func())
	if !ok {
		return
	}
	deleteSession()
}
