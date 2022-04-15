package httpServer

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"net/http"
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
	mux.HandleFunc("/play", playground.Handler("Graphql playground", "/"))
	gqlServer := gqlHandler.NewDefaultServer(es)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("Headers: %+v", req.Cookies())
		_req := req
		for _, middleware := range middlewares {
			_req = middleware(w, req)
		}
		gqlServer.ServeHTTP(w, _req)
	})
}
