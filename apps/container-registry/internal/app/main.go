package app

import (
	"context"
	"fmt"
	"log"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/kloudlite/container-registry-authorizer/auth"
	"go.uber.org/fx"
	"kloudlite.io/apps/container-registry/internal/app/graph"
	"kloudlite.io/apps/container-registry/internal/app/graph/generated"
	"kloudlite.io/apps/container-registry/internal/domain"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type AuthCacheClient cache.Client
type IAMGrpcClient grpc.Client
type EventListnerHttpServer *fiber.App
type AuthorizerHttpServer *fiber.App

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.Repository]("repositories", "prj", entities.RepositoryIndexes),
	repos.NewFxMongoRepo[*entities.Credential]("credentials", "cred", entities.CredentialIndexes),
	repos.NewFxMongoRepo[*entities.Tag]("tags", "tag", entities.TagIndexes),

	fx.Provide(
		func(conn IAMGrpcClient) iam.IAMClient {
			return iam.NewIAMClient(conn)
		},
	),

	fx.Invoke(
		func(
			server *fiber.App,
			d domain.Domain,
			cacheClient AuthCacheClient,
			ev *env.Env,
		) {
			gqlConfig := generated.Config{Resolvers: &graph.Resolver{Domain: d}}
			gqlConfig.Directives.IsLoggedInAndVerified = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				if !sess.UserVerified {
					return nil, &fiber.Error{
						Code:    fiber.StatusForbidden,
						Message: "user's email is not verified",
					}
				}

				return next(context.WithValue(ctx, "user-session", sess))
			}

			gqlConfig.Directives.HasAccount = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}
				m := httpServer.GetHttpCookies(ctx)
				klAccount := m[ev.AccountCookieName]
				if klAccount == "" {
					return nil, fmt.Errorf("no cookie named %q present in request", ev.AccountCookieName)
				}

				nctx := context.WithValue(ctx, "user-session", sess)
				nctx = context.WithValue(nctx, "account-name", klAccount)
				return next(nctx)
			}

			schema := generated.NewExecutableSchema(gqlConfig)
			httpServer.SetupGQLServer(
				server,
				schema,
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					"hotspot-session",
					ev.CookieDomain,
					constants.CacheSessionPrefix,
				),
			)
		},
	),

	fx.Invoke(func(authorizerHttpServer AuthorizerHttpServer, envs *env.Env, d domain.Domain) {
		var a *fiber.App
		a = authorizerHttpServer

		a.Use("/*", func(c *fiber.Ctx) error {

			path := c.Query("path", "/")
			method := c.Query("method", "GET")

			b_auth := basicauth.New(basicauth.Config{
				Realm: "Forbidden",
				Authorizer: func(u string, p string) bool {

					userName, accountName, _, err := auth.ParseToken(p)

					if err != nil {
						log.Println(err)
						return false
					}

					s, err := d.GetTokenKey(c.Context(), userName, accountName)
					if err != nil {
						log.Println(err)
						return false
					}

					if err := auth.Authorizer(u, p, path, method, envs.RegistrySecretKey+s); err != nil {
						log.Println(err)
						return false
					}
					return true
				},
			})

			return b_auth(c)
		})

		a.Get("/*", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})
	}),

	fx.Invoke(func(eventListnerHttpServer EventListnerHttpServer, d domain.Domain) {
		var a *fiber.App
		a = eventListnerHttpServer

		a.Post("/*", func(c *fiber.Ctx) error {

			ctx := c.Context()

			var eventMessage entities.EventMessage
			if err := c.BodyParser(&eventMessage); err != nil {
				return c.SendStatus(400)
			}

			if err := d.ProcessEvents(ctx, eventMessage.Events); err != nil {
				log.Println(err)
				return c.SendStatus(400)
			}

			return c.SendStatus(200)
		})
	}),

	domain.Module,
)
