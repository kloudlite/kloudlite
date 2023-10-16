package app

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	registryAuth "github.com/kloudlite/container-registry-authorizer/auth"
	"go.uber.org/fx"
	"kloudlite.io/apps/container-registry/internal/app/graph"
	"kloudlite.io/apps/container-registry/internal/app/graph/generated"
	"kloudlite.io/apps/container-registry/internal/domain"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
)

type AuthCacheClient cache.Client
type IAMGrpcClient grpc.Client
type AuthGrpcClient grpc.Client
type AuthorizerHttpServer *fiber.App

type venv struct {
	ev *env.Env
}

func (venv *venv) GithubConfig() (clientId, clientSecret, callbackUrl, ghAppId, ghAppPKFile string) {

	return venv.ev.GithubClientId, venv.ev.GithubClientSecret, venv.ev.GithubCallbackUrl, venv.ev.GithubAppId, venv.ev.GithubAppPKFile

}

func (fm *venv) GithubScopes() string {
	return fm.ev.GithubScopes
}

func (fm *venv) GetSubscriptionTopics() []string {
	return []string{fm.ev.KafkaGitWebhookTopics}
}
func (fm *venv) GetConsumerGroupId() string {
	return fm.ev.KafkaConsumerGroup
}

func (fm *venv) GithubWebhookAuthzSecret() string {
	return fm.ev.GithubWebhookAuthzSecret
}

func (fm *venv) GitlabConfig() (clientId, clientSecret, callbackUrl string) {
	return fm.ev.GitlabClientId, fm.ev.GitlabClientSecret, fm.ev.GitlabCallbackUrl
}

func (fm *venv) GitlabScopes() string {
	return fm.ev.GitlabScopes
}

func (fm *venv) GetBrokerHosts() string {
	return fm.ev.KafkaBrokers
}
func (fm *venv) GetKafkaSASLAuth() *redpanda.KafkaSASLAuth {
	return nil
}

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.Repository]("repositories", "prj", entities.RepositoryIndexes),
	repos.NewFxMongoRepo[*entities.Credential]("credentials", "cred", entities.CredentialIndexes),
	repos.NewFxMongoRepo[*entities.Tag]("tags", "tag", entities.TagIndexes),
	repos.NewFxMongoRepo[*entities.Build]("builds", "build", entities.BuildIndexes),

	redpanda.NewConsumerFx[*venv](),
	redpanda.NewProducerFx[*venv](),

	fx.Provide(
		func(conn IAMGrpcClient) iam.IAMClient {
			return iam.NewIAMClient(conn)
		},
	),

	fx.Provide(
		func(conn AuthGrpcClient) auth.AuthClient {
			return auth.NewAuthClient(conn)
		},
	),

	fx.Provide(func(ev *env.Env) *venv {
		return &venv{ev}
	}),

	fxGithub[*venv](),
	fxInvokeProcessGitWebhooks(),

	fxGitlab[*venv](),

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

		a.Post("/events", func(c *fiber.Ctx) error {

			ctx := c.Context()

			var eventMessage entities.EventMessage
			if err := c.BodyParser(&eventMessage); err != nil {
				return c.SendStatus(400)
			}

			if err := d.ProcessRegistryEvents(ctx, eventMessage.Events); err != nil {
				log.Println(err)
				return c.SendStatus(400)
			}

			return c.SendStatus(200)
		})

		a.Use("/auth", func(c *fiber.Ctx) error {

			path := c.Query("path", "/")
			method := c.Query("method", "GET")

			u, err := url.Parse("http://example.com" + path)
			if err != nil {
				return c.SendStatus(400)
			}

			if u.Query().Has("_state") {
				return c.Next()
			}

			b_auth := basicauth.New(basicauth.Config{
				Realm: "Forbidden",
				Authorizer: func(u string, p string) bool {
					if method == "DELETE" && u != domain.KL_ADMIN {
						return false
					}

					userName, accountName, _, err := registryAuth.ParseToken(p)

					if err != nil {
						log.Println(err)
						return false
					}

					s, err := d.GetTokenKey(context.TODO(), userName, accountName)
					if err != nil {
						log.Println(err)
						return false
					}

					if err := registryAuth.Authorizer(u, p, path, method, envs.RegistrySecretKey+s); err != nil {
						log.Println(err)
						return false
					}
					return true
				},
			})

			return b_auth(c)

		})

		a.Get("/auth", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})
	}),

	domain.Module,
)
