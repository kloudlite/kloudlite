package httpServer

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"go.uber.org/fx"
	"net/http"
	"strings"
	"time"

	"github.com/rs/cors"
	"kloudlite.io/pkg/logger"
)

func Start(ctx context.Context, port uint16, mux http.Handler, corsOpt cors.Options, logger logger.Logger) error {
	errChannel := make(chan error, 1)

	c := cors.New(corsOpt)

	go func() {
		// TODO: find a way for graceful shutdown of server
		errChannel <- http.ListenAndServe(fmt.Sprintf(":%v", port), c.Handler(mux))
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
	mux *http.ServeMux,
	es graphql.ExecutableSchema,
	middlewares ...func(http.ResponseWriter, *http.Request) *http.Request,
) {
	mux.HandleFunc("/play", playground.Handler("Graphql playground", "/query"))
	gqlServer := gqlHandler.NewDefaultServer(es)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		_req := req
		for _, middleware := range middlewares {
			_req = middleware(w, req)
		}
		w.Header().Add("EXAMPLE", "sample")
		gqlServer.ServeHTTP(w, _req)
	})
}

type ServerOptions interface {
	GetHttpPort() uint16
	GetHttpCors() string
}

func NewHttpServerFx[T ServerOptions]() fx.Option {
	return fx.Module("http-server",
		fx.Provide(http.NewServeMux),
		fx.Invoke(func(lf fx.Lifecycle, env T, logger logger.Logger, server *http.ServeMux) {
			lf.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					var corsOpt cors.Options
					if env.GetHttpCors() != "" {
						corsOpt = cors.Options{
							AllowedOrigins:   strings.Split(env.GetHttpCors(), ","),
							AllowCredentials: true,
							AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
						}
					}

					// corsOpt = cors.Options{
					// 	AllowedOrigins:   strings.Split(env.GetHttpCors(), ","),
					// 	AllowCredentials: true,
					// 	AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
					// }

					return Start(ctx, env.GetHttpPort(), server, corsOpt, logger)
				},
			})
		}),
	)
}
