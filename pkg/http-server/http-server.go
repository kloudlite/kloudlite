package httpServer

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/adaptor/v2"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	l "github.com/gofiber/fiber/v2/middleware/logger"
	"kloudlite.io/pkg/logging"
)

type Server interface {
	SetupGraphqlServer(es graphql.ExecutableSchema, middlewares ...fiber.Handler)
	Listen(addr string) error
	Close() error
}

type server struct {
	Logger logging.Logger
	*fiber.App
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
		return fmt.Errorf("could not start server because %v", status.Error())
	case <-ctx.Done():
		// if s.Logger != nil {
		s.Logger.Infof("Http Server started @ (addr: %q)", addr)
		// }
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
	s.App.All("/play", adaptor.HTTPHandler(playground.Handler("GraphQL playground", "/query")))
	gqlServer := gqlHandler.NewDefaultServer(es)
	for _, v := range middlewares {
		s.App.Use(v)
	}
	s.App.All("/query", adaptor.HTTPHandlerFunc(gqlServer.ServeHTTP))
}
