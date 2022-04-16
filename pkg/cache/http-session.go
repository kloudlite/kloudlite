package cache

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"kloudlite.io/pkg/repos"
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
		newContext := r.Context()
		if cookie != nil {
			// TODO handle error
			key := fmt.Sprintf("%v:%v", sessionKeyPrefix, cookie.Value)
			var get any
			get, _ = repo.Get(r.Context(), key)
			// TODO handle error

			if get != nil {
				newContext = context.WithValue(r.Context(), "session", get)
			}
		}
		newContext = context.WithValue(newContext, "set-session", func(session T) {
			err := repo.Set(r.Context(), fmt.Sprintf("%v:%v", sessionKeyPrefix, session.GetId()), session)
			if err != nil {
				fmt.Println("[ERROR]", err)
			}
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
			if cookie != nil {
				repo.Drop(newContext, fmt.Sprintf("%v:%v", sessionKeyPrefix, cookie.Value))
			}
		})
		return r.WithContext(newContext)
	}
}

func GetSession[T repos.Entity](ctx context.Context) T {
	value := ctx.Value("session")
	if value != nil {
		return value.(T)
	}
	var x T
	return x
}

func SetSession[T repos.Entity](ctx context.Context, session T) {
	setSession, ok := ctx.Value("set-session").(func(T))
	if !ok {
		return
	}
	setSession(session)
}

func DeleteSession(ctx context.Context) {
	deleteSession, ok := ctx.Value("delete-session").(func())
	if !ok {
		return
	}
	deleteSession()
}
