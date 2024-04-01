package httpServer

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/repos"
)

const userContextKey = "__local_user_context__"

func NewReadSessionMiddleware(repo kv.Repo[*common.AuthSession], cookieName string, sessionKeyPrefix string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		cookies := map[string]string{}
		ctx.Request().Header.VisitAllCookie(func(key, value []byte) {
			cookies[string(key)] = string(value)
		})

		ctx.SetUserContext(context.WithValue(ctx.UserContext(), "http-cookies", cookies))

		cookieValue := ctx.Cookies(cookieName)

		if cookieValue != "" {
			key := fmt.Sprintf("%s:%s", sessionKeyPrefix, cookieValue)
			sess, err := repo.Get(ctx.Context(), key)
			if err != nil {
				if !repo.ErrKeyNotFound(err) {
					return errors.NewE(err)
				}
			}

			if sess != nil {
				ctx.SetUserContext(context.WithValue(ctx.UserContext(), "session", sess))
			}
		}
		return ctx.Next()
	}
}

func NewSessionMiddleware(
	repo kv.Repo[*common.AuthSession],
	cookieName string,
	cookieDomain string,
	sessionKeyPrefix string,
) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		cookies := map[string]string{}
		ctx.Request().Header.VisitAllCookie(func(key, value []byte) {
			cookies[string(key)] = string(value)
		})

		ctx.SetUserContext(context.WithValue(ctx.UserContext(), "http-cookies", cookies))

		cookieValue := ctx.Cookies(cookieName)

		if cookieValue != "" {
			key := fmt.Sprintf("%s:%s", sessionKeyPrefix, cookieValue)
			var get any
			get, err := repo.Get(ctx.Context(), key)
			if err != nil {
				if !repo.ErrKeyNotFound(err) {
					return errors.NewE(err)
				}
			}

			if get != nil {
				ctx.SetUserContext(context.WithValue(ctx.UserContext(), "session", get))
			}
		}

		ctx.SetUserContext(
			context.WithValue(
				ctx.UserContext(), "set-session", func(session *common.AuthSession) {
					err := repo.Set(ctx.Context(), fmt.Sprintf("%v:%v", sessionKeyPrefix, session.GetId()), session)
					if err != nil {
						fmt.Println("[ERROR]", err)
					}
					ck := &fiber.Cookie{
						Name:        cookieName,
						Value:       string(session.GetId()),
						Path:        "/",
						Domain:      fmt.Sprintf("%v", cookieDomain),
						MaxAge:      0,
						Expires:     time.Time{},
						Secure:      true,
						HTTPOnly:    true,
						SameSite:    fiber.CookieSameSiteNoneMode,
						SessionOnly: false,
					}
					// fmt.Println("ck: ", ck)
					ctx.Cookie(ck)
				},
			),
		)

		ctx.SetUserContext(
			context.WithValue(
				ctx.UserContext(), "delete-session", func() {
					if cookieValue != "" {
						if err := repo.Drop(ctx.Context(), fmt.Sprintf("%v:%v", sessionKeyPrefix, cookieValue)); err != nil {
							fmt.Println("[ERROR]", err)
						}
					}
					ctx.Cookie(&fiber.Cookie{
						Name:     cookieName,
						Value:    "expired",
						Path:     "/",
						Domain:   fmt.Sprintf("%v", cookieDomain),
						Expires:  time.Now().Add(-1 * time.Minute),
						MaxAge:   0,
						Secure:   true,
						HTTPOnly: true,
						// SameSite: http.SameSiteStrictMode,
						SameSite: fiber.CookieSameSiteNoneMode,
					})
				},
			),
		)
		return ctx.Next()
	}
}

func GetHttpCookies(ctx context.Context) map[string]string {
	v := ctx.Value(userContextKey)
	if v == nil {
		return nil
	}

	if userCtx, ok := v.(context.Context); ok {
		if cookies, ok := userCtx.Value("http-cookies").(map[string]string); ok {
			return cookies
		}
	}
	return nil
}

func GetSession[T repos.Entity](ctx context.Context) T {
	if ctx.Value(userContextKey) == nil {
		var x T
		return x
	}

	userContext := ctx.Value(userContextKey).(context.Context)
	value := userContext.Value("session")
	if value != nil {
		return value.(T)
	}
	var x T
	return x
}

func SetSession[T repos.Entity](ctx context.Context, session T) {
	userContext := ctx.Value(userContextKey).(context.Context)
	setSession, ok := userContext.Value("set-session").(func(T))
	if !ok {
		fmt.Println("[ERROR]", "set-session is not a function")
		return
	}
	setSession(session)
}

func DeleteSession(ctx context.Context) {
	userContext := ctx.Value(userContextKey).(context.Context)
	deleteSession, ok := userContext.Value("delete-session").(func())
	if !ok {
		return
	}
	deleteSession()
}
