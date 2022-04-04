package framework

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/codingconcepts/env"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"kloudlite.io/apps/wireguard/internal/app"
	"kloudlite.io/apps/wireguard/internal/domain"
	"kloudlite.io/pkg/errors"
)

type Logger struct {
	*zap.SugaredLogger
}

type Env struct {
	MongoUri    string `env:"MONGO_URI", required:"true"`
	MongoDbName string `env:"MONGO_DB_NAME", required:"true"`
	Port        uint32 `env:"PORT", required:"true"`
	IsDev       bool   `env:"IS_DEV", required:"true"`
}

func getEnv(logger Logger) Env {
	var envC Env
	if err := env.Set(&envC); err != nil {
		panic(err)
	}

	isDev := flag.Bool("dev", false, "isDevelopment")
	flag.Parse()
	envC.IsDev = *isDev
	logger.Debugf("%+v\n", envC)
	return envC
}

func NewLogger() Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return Logger{SugaredLogger: logger.Sugar()}
}

func NewMongoDatabase(env Env) (db *mongo.Database, e error) {
	defer errors.HandleErr(&e)
	client, e := mongo.NewClient(options.Client().ApplyURI(env.MongoUri))
	errors.AssertNoError(e, fmt.Errorf("could not create mongo client"))
	e = client.Connect(context.Background())
	errors.AssertNoError(e, fmt.Errorf("could not connect to mongo"))
	return client.Database(env.MongoDbName), nil
}

func NewFramework() FM {
	app := fx.New(fx.Options(
		fx.Provide(NewLogger),
		fx.Provide(getEnv),
		fx.Provide(NewMongoDatabase),
		app.Module,
		fx.Invoke(func(lf fx.Lifecycle, env Env, d domain.Domain, logger Logger, gqlHandler http.Handler) {
			lf.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					errChannel := make(chan error, 1)
					go func() {
						errChannel <- http.ListenAndServe(fmt.Sprintf(":%v", env.Port), gqlHandler)
					}()

					ctx, cancel := context.WithTimeout(ctx, time.Second*1)
					defer cancel()

					select {
					case status := <-errChannel:
						return fmt.Errorf("could not start server because %v", status.Error())
					case <-ctx.Done():
						logger.Infof("Graphql Server started @ (port=%v)", env.Port)
					}
					return nil
				},
				OnStop: func(context.Context) error {
					return nil
				},
			})
		}),
	))

	return func() error {
		app.Run()
		return nil
	}
}
