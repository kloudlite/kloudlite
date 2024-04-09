package app

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/api/apps/iot-console/internal/app/graph"
	"github.com/kloudlite/api/apps/iot-console/internal/app/graph/generated"
	"github.com/kloudlite/api/apps/iot-console/internal/domain"
	"github.com/kloudlite/api/apps/iot-console/internal/entities"
	"github.com/kloudlite/api/apps/iot-console/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/pkg/errors"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
)

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.IOTProject]("projects", "prj", entities.IOTProjectIndexes),
	repos.NewFxMongoRepo[*entities.IOTEnvironment]("environments", "env", entities.IOTEnvironmentIndexes),
	repos.NewFxMongoRepo[*entities.IOTDeployment]("deployments", "depl", entities.IOTDeploymentIndexes),
	repos.NewFxMongoRepo[*entities.IOTDevice]("devices", "dev", entities.IOTDeviceIndexes),
	repos.NewFxMongoRepo[*entities.IOTDeviceBlueprint]("device_blueprints", "devblueprint", entities.IOTDeviceBlueprintIndexes),
	repos.NewFxMongoRepo[*entities.IOTApp]("apps", "app", entities.IOTAppIndexes),

	fx.Invoke(
		func(server httpServer.Server, d domain.Domain, sessionRepo kv.Repo[*common.AuthSession], ev *env.Env) {
			gqlConfig := generated.Config{Resolvers: &graph.Resolver{Domain: d}}

			gqlConfig.Directives.IsLoggedIn = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				return next(context.WithValue(ctx, "user-session", sess))
			}

			gqlConfig.Directives.IsLoggedInAndVerified = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {

				cookies := httpServer.GetHttpCookies(ctx)
				fmt.Printf("cookies: %#v\n", cookies)
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
					return nil, errors.Newf("no cookie named %q present in request", ev.AccountCookieName)
				}

				nctx := context.WithValue(ctx, "user-session", sess)
				nctx = context.WithValue(nctx, "account-name", klAccount)
				return next(nctx)
			}

			schema := generated.NewExecutableSchema(gqlConfig)
			server.SetupGraphqlServer(schema,
				httpServer.NewReadSessionMiddleware(sessionRepo, constants.CookieName, constants.CacheSessionPrefix),
			)
		},
	),

	domain.Module,
)
