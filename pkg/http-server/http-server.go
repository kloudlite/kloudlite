package httpServer

import (
	"context"
	"fmt"
	"github.com/kloudlite/api/pkg/errors"
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
	"github.com/kloudlite/api/pkg/logging"
)

type Server interface {
	SetupGraphqlServer(es graphql.ExecutableSchema, middlewares ...fiber.Handler)
	Listen(addr string) error
	Close() error

	Raw() *fiber.App
}

type server struct {
	Logger logging.Logger
	*fiber.App
}

func (s *server) Raw() *fiber.App {
	return s.App
}

func (s *server) Close() error {
	return s.App.Shutdown()
}

func (s *server) Listen(addr string) error {
	errChannel := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*1)
	defer cancel()

	go func() {
		errChannel <- s.App.Listen(addr)
	}()

	select {
	case status := <-errChannel:
		return errors.Newf("could not start server because %v", status.Error())
	case <-ctx.Done():
		s.Logger.Infof("Http Server started @ (addr: %q)", addr)
	}
	return nil
}

type ServerArgs struct {
	Logger           logging.Logger
	CorsAllowOrigins *string
}

func NewServer(args ServerArgs) Server {
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

	if args.CorsAllowOrigins != nil {
		app.Use(
			cors.New(
				cors.Config{
					AllowOrigins:     *args.CorsAllowOrigins,
					AllowCredentials: true,
					AllowMethods: strings.Join(
						[]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions},
						",",
					),
				},
			),
		)
	}

	if args.Logger == nil {
		args.Logger = logging.EmptyLogger
	}

	return &server{App: app, Logger: args.Logger}
}

func (s *server) SetupGraphqlServer(es graphql.ExecutableSchema, middlewares ...fiber.Handler) {
	s.All("/explorer", func(c *fiber.Ctx) error {
		return c.Redirect(fmt.Sprintf("https://studio.apollographql.com/sandbox/explorer?endpoint=http://%s/query", c.Context().LocalAddr()))
	})
	s.All("/play", adaptor.HTTPHandler(playground.Handler("GraphQL playground", "/query")))
	gqlServer := gqlHandler.NewDefaultServer(es)
	for _, v := range middlewares {
		s.Use(v)
	}
	s.All("/query", adaptor.HTTPHandlerFunc(gqlServer.ServeHTTP))
}
