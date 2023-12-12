package httpServer

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/api/pkg/cache"
	"github.com/kloudlite/api/pkg/repos"
)

const userContextKey = "__local_user_context__"

func NewSessionMiddleware[T repos.Entity](
	cacheClient cache.Client,
	cookieName string,
	cookieDomain string,
	sessionKeyPrefix string,
) fiber.Handler {
	repo := cache.NewRepo[T](cacheClient)
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
				if !repo.ErrNoRecord(err) {
					return err
				}
			}

			if get != nil {
				ctx.SetUserContext(context.WithValue(ctx.UserContext(), "session", get))
			}
		}

		ctx.SetUserContext(
			context.WithValue(
				ctx.UserContext(), "set-session", func(session T) {
					err := repo.Set(ctx.Context(), fmt.Sprintf("%v:%v", sessionKeyPrefix, session.GetId()), session)
					if err != nil {
						fmt.Println("[ERROR]", err)
					}
					ck := &fiber.Cookie{
						Name:     cookieName,
						Value:    string(session.GetId()),
						Path:     "/",
						Domain:   fmt.Sprintf("%v", cookieDomain),
						Expires:  time.Time{},
						MaxAge:   0,
						Secure:   true,
						HTTPOnly: true,
						// SameSite: http.SameSiteStrictMode,
						SameSite: fiber.CookieSameSiteNoneMode,
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
						repo.Drop(ctx.Context(), fmt.Sprintf("%v:%v", sessionKeyPrefix, cookieValue))
					}
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
	deleteSession, ok := ctx.Value("delete-session").(func())
	if !ok {
		return
	}
	deleteSession()
}
