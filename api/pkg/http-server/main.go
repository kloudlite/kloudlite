package httpServer

import (
	"context"
	"fmt"
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
