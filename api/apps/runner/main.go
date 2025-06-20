package main

import (
	"context"
	"flag"
	"fmt"
	accountsapp "github.com/kloudlite/api/apps/accounts/fx-app"
	authapp "github.com/kloudlite/api/apps/auth/fx-app"
	commsapp "github.com/kloudlite/api/apps/comms/fx-app"
	"github.com/kloudlite/api/apps/infra/protobufs/infra"
	"github.com/kloudlite/api/apps/runner/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
	"github.com/kloudlite/api/pkg/repos"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net"
	"os"
	"time"
)

func main() {
	common.PrintBuildInfo()
	start := time.Now()
	var isDev bool
	var debug bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.BoolVar(&debug, "debug", false, "--debug")
	flag.Parse()
	logger := logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, ShowDebugLogs: debug, SetAsDefaultLogger: true})
	app := fx.New(
		// ENV
		fx.Provide(func() (*env.Env, error) {
			if e, err := env.LoadEnv(); err != nil {
				return nil, errors.NewE(err)
			} else {
				e.IsDev = isDev
				return e, nil
			}
		}),

		// GRPC Clients
		fx.Module(
			"grpc-clients",
			fx.Provide(func(env *env.Env) (auth.AuthClient, error) {
				conn, err := grpc.NewClient(env.CommsServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					return nil, errors.NewE(err)
				}
				return auth.NewAuthClient(conn), nil
			}),
			fx.Provide(func(env *env.Env) (auth.AuthInternalClient, error) {
				conn, err := grpc.NewClient(env.CommsServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					return nil, errors.NewE(err)
				}
				return auth.NewAuthInternalClient(conn), nil
			}),
			fx.Provide(func(env *env.Env) (console.ConsoleClient, error) {
				conn, err := grpc.NewClient(env.ConsoleServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					return nil, errors.NewE(err)
				}
				return console.NewConsoleClient(conn), nil
			}),
			fx.Provide(func(env *env.Env) (infra.InfraClient, error) {
				conn, err := grpc.NewClient(env.InfraServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					return nil, errors.NewE(err)
				}
				return infra.NewInfraClient(conn), nil
			}),
			fx.Provide(func(env *env.Env) (iam.IAMClient, error) {
				conn, err := grpc.NewClient(env.IAMServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					return nil, errors.NewE(err)
				}
				return iam.NewIAMClient(conn), nil
			}),
			fx.Provide(func(env *env.Env) (comms.CommsClient, error) {
				conn, err := grpc.NewClient(env.CommsServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					return nil, errors.NewE(err)
				}
				return comms.NewCommsClient(conn), nil
			}),
		),

		// Logger
		fx.Provide(func() *slog.Logger {
			if debug {
				return logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, ShowDebugLogs: true, SetAsDefaultLogger: true})
			}
			return logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, SetAsDefaultLogger: true})
		}),

		fx.Module(
			"grpc-server",
			fx.Provide(func() *grpc.Server {
				return grpc.NewServer()
			}),
			fx.Invoke(func(lifecycle fx.Lifecycle, server *grpc.Server, env *env.Env, logger *slog.Logger) {
				lifecycle.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							listener, err := net.Listen("tcp", fmt.Sprintf(":%d", env.GrpcPort))
							if err != nil {
								return errors.NewEf(err, "could not listen on port %d", env.GrpcPort)
							}
							logger.Info(fmt.Sprintf("Starting gRPC server on port %d", env.GrpcPort))
							go server.Serve(listener)
							return nil
						},
						OnStop: func(ctx context.Context) error {
							logger.Info(fmt.Sprintf("Stopping gRPC server on port %d", env.GrpcPort))
							server.Stop()
							return nil
						},
					},
				)
			}),
		),

		fx.Module(
			"nats-js-client",
			fx.Provide(func(ev *env.Env, logger *slog.Logger) (*nats.Client, error) {
				return nats.NewClient(ev.NatsURL, nats.ClientOpts{
					Name:   "kl",
					Logger: logger,
				})
			}),
			fx.Provide(func(ev *env.Env, logger *slog.Logger) (*nats.JetstreamClient, error) {
				name := "auth:jetstream-client"
				nc, err := nats.NewClient(ev.NatsURL, nats.ClientOpts{
					Name:   name,
					Logger: logger,
				})
				if err != nil {
					return nil, errors.NewE(err)
				}
				return nats.NewJetstreamClient(nc)
			}),
		),

		fx.Module(
			"mongo-connection",
			fx.Provide(func(env *env.Env) (*mongo.Database, error) {
				ctx, cf := context.WithTimeout(context.TODO(), 10*time.Second)
				defer cf()
				return repos.NewMongoDatabase(ctx, "mongodb://localhost:27017", "kloudlite")
			}),
			fx.Invoke(func(db *mongo.Database, lifecycle fx.Lifecycle) {
				lifecycle.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						if err := db.Client().Ping(ctx, nil); err != nil {
							return errors.NewEf(err, "could not ping Mongo")
						}
						slog.Info("connected to mongodb database", "db", db.Name())
						return nil
					},
					OnStop: func(ctx context.Context) error {
						return db.Client().Disconnect(ctx)
					},
				})
			}),
		),

		fn.FxErrorHandler(),

		authapp.NewAuthModule(),
		commsapp.NewCommsModule(),
		accountsapp.NewAccountsModule(),
	)

	ctx, cancelFunc := func() (context.Context, context.CancelFunc) {
		if isDev {
			return context.WithTimeout(context.TODO(), 10*time.Second)
		}
		return context.WithTimeout(context.TODO(), 5*time.Second)
	}()

	defer cancelFunc()

	if err := app.Start(ctx); err != nil {
		logger.Error("failed to start auth-api, got", "err", err)
		os.Exit(1)
	}

	common.PrintReadyBanner2(time.Since(start))
	<-app.Done()
}
