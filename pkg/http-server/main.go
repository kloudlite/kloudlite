package httpServer

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/fx"
	"kloudlite.io/pkg/logging"
	"net/http"
	"strings"
	"time"
)

func Start(ctx context.Context, port uint16, app *fiber.App, logger logging.Logger) error {
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
		logger.Infof("Graphql Server started @ (port=%v)", port)
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

func NewHttpServerFx[T ServerOptions]() fx.Option {
	return fx.Module(
		"http-server",
		fx.Provide(
			func() *fiber.App {
				return fiber.New()
			},
		),
		fx.Invoke(
			func(lf fx.Lifecycle, env T, logger logging.Logger, app *fiber.App) {
				if env.GetHttpCors() != "" {
					app.Use(
						cors.New(
							cors.Config{
								AllowOrigins:     env.GetHttpCors(),
								AllowCredentials: true,
								AllowMethods: strings.Join(
									[]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions},
									",",
								),
							},
						),
					)
				}

				lf.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							return Start(ctx, env.GetHttpPort(), app, logger)
						},
					},
				)
			},
		),
	)
}
