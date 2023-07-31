package httpServer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	l "github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/fx"
	"kloudlite.io/pkg/logging"
)

func start(ctx context.Context, port uint16, app *fiber.App, logger logging.Logger) error {
	errChannel := make(chan error, 1)
	go func() {
		errChannel <- app.Listen(fmt.Sprintf(":%d", port))
	}()

	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	select {
	case status := <-errChannel:
		return fmt.Errorf("could not start server because %v", status.Error())
	case <-ctx.Done():
		logger.Infof("Http Server started @ (port=%v)", port)
	}
	return nil
}

func SetupGQLServer(
	app *fiber.App,
	es graphql.ExecutableSchema,
	middlewares ...fiber.Handler,
) {
	app.All("/play", adaptor.HTTPHandler(playground.Handler("GraphQL playground", "/query")))
	gqlServer := gqlHandler.NewDefaultServer(es)
	for _, v := range middlewares {
		app.Use(v)
	}
	app.All("/query", adaptor.HTTPHandlerFunc(gqlServer.ServeHTTP))
}

type ServerOptions interface {
	GetHttpPort() uint16
	GetHttpCors() string
}

type HttpServerV2Opts struct {
	CorsOrigins *string
}

func NewHttpServerV2[T ~*fiber.App](opts HttpServerV2Opts) T {
	app := fiber.New()
	app.Use(
		l.New(
			l.Config{
				Format:     "${time} ${status} - ${method} ${latency} \t ${path} \n",
				TimeFormat: "02-Jan-2006 15:04:05",
				TimeZone:   "Asia/Kolkata",
			},
		),
	)

	if opts.CorsOrigins != nil {
		app.Use(
			cors.New(
				cors.Config{
					AllowOrigins:     *opts.CorsOrigins,
					AllowCredentials: true,
					AllowMethods: strings.Join(
						[]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions},
						",",
					),
				},
			),
		)
	}

	return app
}

func StartHttpServerV2[T ~*fiber.App](ctx context.Context, server T, port uint16, logger logging.Logger) error {
	return start(context.Background(), port, server, logger)
}

func StopHttpServerV2[T ~*fiber.App](server T) error {
	return (*fiber.App)(server).Shutdown()
}

func NewHttpServerFx[T ServerOptions]() fx.Option {
	return fx.Module(
		"http-server",
		fx.Provide(
			func(serverOpts T) *fiber.App {
				app := fiber.New()

				app.Use(
					l.New(
						l.Config{
							Format:     "${time} ${status} - ${method} ${latency} \t ${path} \n",
							TimeFormat: "02-Jan-2006 15:04:05",
							TimeZone:   "Asia/Kolkata",
						},
					),
				)

				app.Use(
					cors.New(
						cors.Config{
							AllowOrigins:     serverOpts.GetHttpCors(),
							AllowCredentials: true,
							AllowMethods: strings.Join(
								[]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions},
								",",
							),
						},
					),
				)

				return app
			},
		),

		fx.Invoke(
			func(lf fx.Lifecycle, env T, logger logging.Logger, app *fiber.App) {
				lf.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							return start(ctx, env.GetHttpPort(), app, logger)
						},
						OnStop: func(ctx context.Context) error {
							return app.Shutdown()
						},
					},
				)
			},
		),
	)
}
