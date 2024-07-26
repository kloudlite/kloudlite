package httpServer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kloudlite/api/pkg/errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/ztrue/tracerr"

	"github.com/99designs/gqlgen/graphql"
	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/adaptor/v2"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	l "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/skip"
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
	isDev bool
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

	s.App.Get("/_healthy", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

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
	IsDev            bool
	Logger           logging.Logger
	CorsAllowOrigins *string
	IAMGrpcAddr      string `env:"IAM_GRPC_ADDR" required:"true"`
}

func NewServer(args ServerArgs) Server {
	app := fiber.New(fiber.Config{
		// Prefork:               true,
		DisableStartupMessage: true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			args.Logger.Errorf(err)
			return errors.NewE(err)
		},
	})

	loggerMiddleware := l.New(
		l.Config{
			CustomTags:    map[string]l.LogFunc{},
			Format:        "${time} ${status} - ${method} ${latency} \t ${path} \n",
			TimeFormat:    "02-Jan-2006 15:04:05",
			TimeZone:      "Asia/Kolkata",
			TimeInterval:  0,
			Output:        nil,
			DisableColors: false,
		},
	)

	app.Use(skip.New(loggerMiddleware, func(c *fiber.Ctx) bool {
		return c.Path() == "/_healthy"
	}))

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

	return &server{App: app, Logger: args.Logger, isDev: args.IsDev}
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

	gqlServer.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		if s.isDev {
			tracerr.Print(err.(*gqlerror.Error).Unwrap())
		}
		return gqlerror.Errorf(err.Error())
	})

	s.All("/query", adaptor.HTTPHandlerFunc(gqlServer.ServeHTTP))

	s.All("/", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusNotFound)
	})
}
