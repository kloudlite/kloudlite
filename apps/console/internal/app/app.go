package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	fWebsocket "github.com/gofiber/websocket/v2"
	"kloudlite.io/apps/console/internal/app/graph"
	"kloudlite.io/apps/console/internal/app/graph/generated"
	domain "kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/entities"
	"kloudlite.io/apps/console/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/infra"
	"kloudlite.io/pkg/cache"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/kafka"
	"kloudlite.io/pkg/logging"
	loki_client "kloudlite.io/pkg/loki-client"
	"kloudlite.io/pkg/repos"
)

type (
	AuthCacheClient          cache.Client
	IAMGrpcClient            grpc.Client
	InfraClient              grpc.Client
	LogsAndMetricsHttpServer httpServer.Server
)

func toConsoleContext(requestCtx context.Context, accountCookieName string, clusterCookieName string) (domain.ConsoleContext, error) {
	sess := httpServer.GetSession[*common.AuthSession](requestCtx)
	if sess == nil {
		return domain.ConsoleContext{}, fiber.ErrUnauthorized
	}
	m := httpServer.GetHttpCookies(requestCtx)
	klAccount := m[accountCookieName]
	if klAccount == "" {
		return domain.ConsoleContext{}, fmt.Errorf("no cookie named '%s' present in request", accountCookieName)
	}

	klCluster := m[clusterCookieName]
	if klCluster == "" {
		return domain.ConsoleContext{}, fmt.Errorf("no cookie named '%s' present in request", clusterCookieName)
	}

	return domain.NewConsoleContext(requestCtx, sess.UserId, klAccount, klCluster), nil
}

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.Project]("projects", "prj", entities.ProjectIndexes),
	repos.NewFxMongoRepo[*entities.Workspace]("workspaces", "ws", entities.WorkspaceIndexes),
	repos.NewFxMongoRepo[*entities.App]("apps", "app", entities.AppIndexes),
	repos.NewFxMongoRepo[*entities.Config]("configs", "cfg", entities.ConfigIndexes),
	repos.NewFxMongoRepo[*entities.Secret]("secrets", "scrt", entities.SecretIndexes),
	repos.NewFxMongoRepo[*entities.ManagedResource]("managed_resources", "mres", entities.MresIndexes),
	repos.NewFxMongoRepo[*entities.ManagedService]("managed_services", "msvc", entities.MsvcIndexes),
	repos.NewFxMongoRepo[*entities.Router]("routers", "rt", entities.RouterIndexes),
	repos.NewFxMongoRepo[*entities.ImagePullSecret]("image_pull_secrets", "ips", entities.ImagePullSecretIndexes),

	// streaming logs
	fx.Invoke(
		func(logAndMetricsServer LogsAndMetricsHttpServer, client loki_client.LokiClient,
			ev *env.Env, cacheClient AuthCacheClient, d domain.Domain, logger logging.Logger,
			infraClient infra.InfraClient,
		) {
			a := logAndMetricsServer.Raw()

			a.Use(
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					constants.CookieName,
					ev.CookieDomain,
					constants.CacheSessionPrefix,
				),
			)

			a.Use("/observability", func(c *fiber.Ctx) error {
				cc, err := toConsoleContext(c.Context(), ev.AccountCookieName, ev.ClusterCookieName)
				if err != nil {
					return err
				}

				st := c.Query("start_time")
				et := c.Query("end_time")

				var startTime *time.Time
				var endTime *time.Time

				if st != "" {
					st, err := strconv.ParseInt(st, 10, 64)
					if err != nil {
						return err
					}
					startTime = fn.New(time.Unix(st, 0))
				}

				if et != "" {
					et, err := strconv.ParseInt(et, 10, 64)
					if err != nil {
						return err
					}
					endTime = fn.New(time.Unix(et, 0))
				}

				args := ObservabilityArgs{
					AccountName: cc.AccountName,
					ClusterName: cc.ClusterName,

					ResourceName:      c.Query("resource_name"),
					ResourceNamespace: c.Query("resource_namespace"),
					ResourceType:      c.Query("resource_type"),
					WorkspaceName:     c.Query("workspace_name"),
					ProjectName:       c.Query("project_name"),

					JobName:      c.Query("job_name"),
					JobNamespace: c.Query("job_namespace"),

					StartTime: startTime,
					EndTime:   endTime,
				}

				if b, err := args.Validate(); !b {
					return err
				}

				c.Locals("observability-args", args)
				c.Locals("console-context", cc)
				return c.Next()
			})

			a.Get("/observability/logs/:resource_type",
				func(c *fiber.Ctx) error {
					args, ok := c.Locals("observability-args").(ObservabilityArgs)
					if !ok {
						return fiber.ErrInternalServerError
					}

					cc, ok := c.Locals("console-context").(domain.ConsoleContext)
					if !ok {
						return fiber.ErrInternalServerError
					}

					streamSelectors := make([]loki_client.StreamSelector, 0, 5)

					streamSelectors = append(streamSelectors,
						loki_client.StreamSelector{
							Key:       "kl_account_name",
							Operation: "=",
							Value:     cc.AccountName,
						},
						loki_client.StreamSelector{
							Key:       "kl_cluster_name",
							Operation: "=",
							Value:     cc.ClusterName,
						},
					)

					resourceType := c.Params("resource_type")
					switch resourceType {
					case "app":
						{
							app, err := d.GetApp(cc, args.ResourceNamespace, args.ResourceName)
							if err != nil {
								return err
							}
							logger.Infof("userId: %s, has access to resource 'app': %s/%s, allowing user to consume logs", cc.UserId, app.Namespace, app.Name)

							streamSelectors = append(streamSelectors,
								loki_client.StreamSelector{
									Key:       "kl_resource_name",
									Operation: "=",
									Value:     args.ResourceName,
								},
								loki_client.StreamSelector{
									Key:       "kl_resource_namespace",
									Operation: "=",
									Value:     args.ResourceNamespace,
								},
							)
						}
					case "cluster-job":
						{
							out, err := infraClient.GetCluster(cc, &infra.GetClusterIn{
								UserId:      string(cc.UserId),
								UserName:    cc.UserName,
								UserEmail:   cc.UserEmail,
								AccountName: cc.AccountName,
								ClusterName: cc.ClusterName,
							})
							if err != nil {
								return err
							}
							logger.Infof("userId: %s, has access to resource 'cluster': account: %s, cluster: %s, allowing user to consume logs", cc.UserId, cc.AccountName, cc.ClusterName)
							streamSelectors = nil
							streamSelectors = append(streamSelectors,
								loki_client.StreamSelector{
									Key:       "kl_job_name",
									Operation: "=",
									Value:     out.IACJobName,
								},
								loki_client.StreamSelector{
									Key:       "kl_job_namespace",
									Operation: "=",
									Value:     out.IACJobNamespace,
								},
							)
						}
					case "nodepool-job":
						{
							out, err := infraClient.GetNodepool(cc, &infra.GetNodepoolIn{
								UserId:       string(cc.UserId),
								UserName:     cc.UserName,
								UserEmail:    cc.UserEmail,
								AccountName:  cc.AccountName,
								ClusterName:  cc.ClusterName,
								NodepoolName: args.ResourceName,
							})
							if err != nil {
								return err
							}
							logger.Infof("userId: %s, has access to resource 'nodepool': account: %s, cluster: %s, nodepool: %s,  allowing user to consume logs", cc.UserId, cc.AccountName, cc.ClusterName, args.ResourceName)
							streamSelectors = append(streamSelectors,
								loki_client.StreamSelector{
									Key:       "kl_job_name",
									Operation: "=",
									Value:     out.IACJobName,
								},
								loki_client.StreamSelector{
									Key:       "kl_job_namespace",
									Operation: "=",
									Value:     out.IACJobNamespace,
								},
							)
						}
					default:
						{
							return &fiber.Error{Code: fiber.ErrBadRequest.Code, Message: fmt.Sprintf("unknown resource type (%s), must be on e of app,cluster-job,nodepool-job", resourceType)}
						}
					}

					if args.WorkspaceName != "" {
						streamSelectors = append(streamSelectors, loki_client.StreamSelector{
							Key:       "kl_workspace_name",
							Operation: "=",
							Value:     args.WorkspaceName,
						})
					}

					if args.ProjectName != "" {
						streamSelectors = append(streamSelectors, loki_client.StreamSelector{
							Key:       "kl_project_name",
							Operation: "=",
							Value:     args.ProjectName,
						})
					}

					lokiQueryFilter := &loki_client.QueryArgs{
						StreamSelectors: streamSelectors,
						SearchKeyword:   nil,
						StartTime:       args.StartTime,
						EndTime:         args.EndTime,
						LimitLength:     nil,

						PreWriteFunc: func(lr *loki_client.LogResult) ([]byte, error) {
							var logMessage struct {
								Message string `json:"message"`
							}

							type LogFormat struct {
								Timestamp string `json:"timestamp"`
								Message   string `json:"message"`
							}

							type FinalResult struct {
								PodName string      `json:"pod_name"`
								Logs    []LogFormat `json:"logs"`
							}

							data := make([]FinalResult, len(lr.Data.Result))
							for i := range lr.Data.Result {
								data[i] = FinalResult{
									PodName: lr.Data.Result[i].Stream["kl_pod_name"],
									Logs:    make([]LogFormat, len(lr.Data.Result[i].Values)),
								}

								for j := range lr.Data.Result[i].Values {
									ts, err := strconv.ParseInt(lr.Data.Result[i].Values[j][0], 10, 64)
									if err != nil {
										return nil, err
									}
									data[i].Logs[j].Timestamp = time.Unix(0, ts).Format(time.RFC3339)
									if err := json.Unmarshal([]byte(lr.Data.Result[i].Values[j][1]), &logMessage); err != nil {
										return nil, err
									}
									data[i].Logs[j].Message = logMessage.Message
								}
							}
							return json.Marshal(data)
						},
					}

					c.Locals("loki-query-filter", lokiQueryFilter)
					return c.Next()
				},

				func(c *fiber.Ctx) error {
					if fWebsocket.IsWebSocketUpgrade(c) {
						c.Locals("allowed", true)
						return c.Next()
					}

					lokiQueryFilter, ok := c.Locals("loki-query-filter").(*loki_client.QueryArgs)
					if !ok {
						return fiber.ErrInternalServerError
					}

					b, err := client.GetLogs(*lokiQueryFilter)
					if err != nil {
						return err
					}

					if _, err := c.Write(b); err != nil {
						return err
					}

					return nil
				},

				fWebsocket.New(
					func(conn *fWebsocket.Conn) {
						defer conn.Close()

						pr, pw := io.Pipe()

						go func() {
							// now read from pr, and write it to websocket conn
							defer pr.Close()
							defer conn.Close()

							r := bufio.NewReader(pr)
							msg := make([]byte, 0xffff)
							for {
								n, err := r.Read(msg)
								if err != nil {
									if err != io.EOF {
										conn.WriteMessage(fWebsocket.CloseInternalServerErr, []byte(err.Error()))
										return
									}
									if conn.Conn == nil {
										break
									}
									conn.WriteMessage(fWebsocket.TextMessage, msg[:n])
									return
								}

								conn.WriteMessage(fWebsocket.TextMessage, msg[:n])
							}
						}()

						lokiQueryFilter, ok := conn.Locals("loki-query-filter").(*loki_client.QueryArgs)
						if !ok {
							conn.WriteMessage(fWebsocket.CloseMessage, []byte(fiber.ErrInternalServerError.Error()))
							return
						}

						if err := client.TailLogs(*lokiQueryFilter, pw); err != nil {
							return
						}
					}),
			)

			a.Get("/observability/metrics/:metric_type", func(c *fiber.Ctx) error {
				metricType := c.Params("metric_type")

				args, ok := c.Locals("observability-args").(ObservabilityArgs)
				if !ok {
					return fiber.ErrInternalServerError
				}

				return queryProm(ev.PromHttpAddr, PromMetricsType(metricType), map[ObservabilityLabel]string{
					ResourceName:      args.ResourceName,
					ResourceNamespace: args.ResourceNamespace,
					ResourceType:      args.ResourceType,

					WorkspaceName: args.WorkspaceName,
					ProjectName:   args.ProjectName,
				}, args.StartTime, args.EndTime, c.Response().BodyWriter())
			})
		},
	),

	fx.Invoke(
		func(server httpServer.Server, d domain.Domain, cacheClient AuthCacheClient, ev *env.Env) {
			gqlConfig := generated.Config{Resolvers: &graph.Resolver{Domain: d}}

			gqlConfig.Directives.IsLoggedIn = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				return next(context.WithValue(ctx, "user-session", sess))
			}

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

			gqlConfig.Directives.HasAccountAndCluster = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				m := httpServer.GetHttpCookies(ctx)
				klAccount := m[ev.AccountCookieName]
				if klAccount == "" {
					return nil, fmt.Errorf("no cookie named '%s' present in request", ev.AccountCookieName)
				}

				klCluster := m[ev.ClusterCookieName]
				if klCluster == "" {
					return nil, fmt.Errorf("no cookie named '%s' present in request", ev.ClusterCookieName)
				}

				nctx := context.WithValue(ctx, "user-session", sess)
				nctx = context.WithValue(nctx, "account-name", klAccount)
				nctx = context.WithValue(nctx, "cluster-name", klCluster)
				return next(nctx)
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
			server.SetupGraphqlServer(schema,
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					constants.CookieName,
					ev.CookieDomain,
					constants.CacheSessionPrefix,
				),
			)
		},
	),

	fx.Provide(func(conn kafka.Conn, logger logging.Logger) (domain.MessageDispatcher, error) {
		return kafka.NewProducer(conn, kafka.ProducerOpts{
			Logger: logger.WithName("message-dispatcher"),
		})
	}),
	fx.Invoke(func(lf fx.Lifecycle, producer domain.MessageDispatcher) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return producer.LifecycleOnStart(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return producer.LifecycleOnStop(ctx)
			},
		})
	}),

	// fx.Invoke(ProcessErrorOnApply),

	fx.Provide(
		func(conn IAMGrpcClient) iam.IAMClient {
			return iam.NewIAMClient(conn)
		},
	),

	fx.Provide(
		func(conn InfraClient) infra.InfraClient {
			return infra.NewInfraClient(conn)
		},
	),

	domain.Module,

	fx.Provide(func(conn kafka.Conn, ev *env.Env, logger logging.Logger) (ErrorOnApplyConsumer, error) {
		return kafka.NewConsumer(conn, ev.KafkaConsumerGroupId, []string{ev.KafkaErrorOnApplyTopic}, kafka.ConsumerOpts{
			Logger: logger.WithName("error-consumer"),
		})
	}),
	fx.Invoke(func(lf fx.Lifecycle, consumer ErrorOnApplyConsumer, d domain.Domain) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				if err := consumer.LifecycleOnStart(ctx); err != nil {
					return err
				}
				go ProcessErrorOnApply(consumer, d)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return consumer.LifecycleOnStop(ctx)
			},
		})
	}),

	fx.Provide(func(conn kafka.Conn, ev *env.Env, logger logging.Logger) (ResourceUpdateConsumer, error) {
		return kafka.NewConsumer(conn, ev.KafkaConsumerGroupId, []string{ev.KafkaStatusUpdatesTopic}, kafka.ConsumerOpts{
			Logger: logger.WithName("resource-update-consumer"),
		})
	}),
	fx.Invoke(func(lf fx.Lifecycle, consumer ResourceUpdateConsumer, d domain.Domain) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				if err := consumer.LifecycleOnStart(ctx); err != nil {
					return err
				}
				go ProcessResourceUpdates(consumer, d)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return consumer.LifecycleOnStop(ctx)
			},
		})
	}),
)
