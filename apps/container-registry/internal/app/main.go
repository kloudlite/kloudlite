package app

import (
	"context"
	"net/url"

	"github.com/kloudlite/api/pkg/cache"
	"github.com/kloudlite/api/pkg/errors"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/nats"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/kloudlite/api/apps/container-registry/internal/app/graph"
	"github.com/kloudlite/api/apps/container-registry/internal/app/graph/generated"
	"github.com/kloudlite/api/apps/container-registry/internal/domain"
	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	"github.com/kloudlite/api/apps/container-registry/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/repos"
	registryAuth "github.com/kloudlite/container-registry-authorizer/auth"
	"go.uber.org/fx"
)

type (
	IAMGrpcClient        grpc.Client
	AuthGrpcClient       grpc.Client
	AuthorizerHttpServer httpServer.Server
)

type venv struct {
	ev *env.Env
}

func (venv *venv) GithubConfig() (clientId, clientSecret, callbackUrl, ghAppId, ghAppPKFile string) {
	return venv.ev.GithubClientId, venv.ev.GithubClientSecret, venv.ev.GithubCallbackUrl, venv.ev.GithubAppId, venv.ev.GithubAppPKFile
}

func (fm *venv) GithubScopes() string {
	return fm.ev.GithubScopes
}

func (fm *venv) GitlabConfig() (clientId, clientSecret, callbackUrl string) {
	return fm.ev.GitlabClientId, fm.ev.GitlabClientSecret, fm.ev.GitlabCallbackUrl
}

func (fm *venv) GitlabScopes() string {
	return fm.ev.GitlabScopes
}

func (fm *venv) GitlabWebhookAuthzSecret() *string {
	return &fm.ev.GitlabWebhookAuthzSecret
}

func (fm *venv) GitlabWebhookUrl() *string {
	return &fm.ev.GitlabWebhookUrl
}

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.Repository]("repositories", "prj", entities.RepositoryIndexes),
	repos.NewFxMongoRepo[*entities.Credential]("credentials", "cred", entities.CredentialIndexes),
	repos.NewFxMongoRepo[*entities.Digest]("tags", "tag", entities.TagIndexes),
	repos.NewFxMongoRepo[*entities.Build]("builds", "build", entities.BuildIndexes),
	repos.NewFxMongoRepo[*entities.BuildCacheKey]("build-caches", "build-cache", entities.BuildCacheKeyIndexes),
	repos.NewFxMongoRepo[*entities.BuildRun]("build_runs", "build_run", entities.BuildRunIndices),

	fx.Provide(func(jc *nats.JetstreamClient, ev *env.Env, logger logging.Logger) (GitWebhookConsumer, error) {
		topic := string(common.GitWebhookTopicName)
		consumerName := "container-reg:git-webhooks"
		return msg_nats.NewJetstreamConsumer(context.TODO(), jc, msg_nats.JetstreamConsumerArgs{
			Stream: ev.NatsStream,
			ConsumerConfig: msg_nats.ConsumerConfig{
				Name:           consumerName,
				Durable:        consumerName,
				Description:    "this consumer reads message from a subject dedicated to errors, that occurred when the resource was applied at the agent",
				FilterSubjects: []string{topic},
			},
		})
	}),

	fx.Provide(func(jc *nats.JetstreamClient, ev *env.Env, logger logging.Logger) BuildRunProducer {
		return msg_nats.NewJetstreamProducer(jc)
	}),

	fx.Provide(func(jsc *nats.JetstreamClient, ev *env.Env) (ReceiveResourceUpdatesConsumer, error) {
		topic := common.GetPlatformClusterMessagingTopic("*", "*", common.ContainerRegistryReceiver, common.EventResourceUpdate)

		consumerName := "cr:resource-updates"
		return msg_nats.NewJetstreamConsumer(context.TODO(), jsc, msg_nats.JetstreamConsumerArgs{
			Stream: ev.NatsStream,
			ConsumerConfig: msg_nats.ConsumerConfig{
				Name:           consumerName,
				Durable:        consumerName,
				Description:    "this consumer receives container registry resource updates, processes them, and keeps our Database updated about things happening in the cluster",
				FilterSubjects: []string{topic},
			},
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, consumer ReceiveResourceUpdatesConsumer, d domain.Domain, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go processResourceUpdates(consumer, d, logger)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return consumer.Stop(ctx)
			},
		})
	}),

	fx.Provide(func(jsc *nats.JetstreamClient, ev *env.Env) (ErrorOnApplyConsumer, error) {
		topic := common.GetPlatformClusterMessagingTopic("*", "*", common.ConsoleReceiver, common.EventErrorOnApply)

		consumerName := "cr:error-on-apply"

		return msg_nats.NewJetstreamConsumer(context.TODO(), jsc, msg_nats.JetstreamConsumerArgs{
			Stream: ev.NatsStream,
			ConsumerConfig: msg_nats.ConsumerConfig{
				Name:           consumerName,
				Durable:        consumerName,
				Description:    "this consumer receives container registry resource apply error on agent, processes them, and keeps our Database updated about why the resource apply failed at agent",
				FilterSubjects: []string{topic},
			},
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, consumer ErrorOnApplyConsumer, d domain.Domain, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go ProcessErrorOnApply(consumer, logger, d)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return consumer.Stop(ctx)
			},
		})
	}),

	fx.Provide(func(jsc *nats.JetstreamClient, logger logging.Logger) SendTargetClusterMessagesProducer {
		return msg_nats.NewJetstreamProducer(jsc)
	}),

	fx.Provide(func(targetMessageProducer SendTargetClusterMessagesProducer) domain.ResourceDispatcher {
		return NewResourceDispatcher(targetMessageProducer)
	}),

	fx.Provide(func(cli *nats.Client, logger logging.Logger) domain.ResourceEventPublisher {
		return NewResourceEventPublisher(cli, logger)
	}),



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
	fxGitlab[*venv](),

	fx.Invoke(
		func(server httpServer.Server, d domain.Domain, sessionRepo cache.Repo[*common.AuthSession], ev *env.Env) {
			gqlConfig := generated.Config{Resolvers: &graph.Resolver{Domain: d}}

			gqlConfig.Directives.IsLoggedInAndVerified = func(ctx context.Context, _ interface{}, next graphql.Resolver) (res interface{}, err error) {
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

			gqlConfig.Directives.HasAccount = func(ctx context.Context, _ interface{}, next graphql.Resolver) (res interface{}, err error) {
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
			server.SetupGraphqlServer(schema, httpServer.NewSessionMiddleware(
				sessionRepo,
				"hotspot-session",
				ev.CookieDomain,
				constants.CacheSessionPrefix,
			))
		},
	),

	fx.Invoke(func(server AuthorizerHttpServer, envs *env.Env, d domain.Domain, logger logging.Logger) {
		a := server.Raw()
		a.Post("/events", func(c *fiber.Ctx) error {
			ctx := c.Context()

			var eventMessage entities.EventMessage
			if err := c.BodyParser(&eventMessage); err != nil {
				return c.SendStatus(400)
			}

			if err := d.ProcessRegistryEvents(ctx, eventMessage.Events, logger); err != nil {
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

			if method == "HEAD" {
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
						return false
					}

					s, err := d.GetTokenKey(context.TODO(), userName, accountName)
					if err != nil {
						return false
					}

					if err := registryAuth.Authorizer(u, p, path, method, envs.RegistrySecretKey+s); err != nil {
						return false
					}
					return true
				},
			})

			r := b_auth(c)

			return r
		})

		a.Get("/auth", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})
	}),

	domain.Module,

	fx.Invoke(func(lf fx.Lifecycle, d domain.Domain, consumer GitWebhookConsumer, producer BuildRunProducer, logr logging.Logger, envs *env.Env) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					err := processGitWebhooks(ctx, d, consumer, producer, logr, envs)
					if err != nil {
						logr.Errorf(err, "could not process git webhooks")
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return nil
			},
		})
	}),
)
