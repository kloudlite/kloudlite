package gql_server

import (
	"context"
	"fmt"
	"kloudlite.io/pkg/logger"
	"net/http"
	"time"
)

func StartGQLServer(ctx context.Context, port uint32, gqlHandler http.Handler, logger logger.Logger) error {
	errChannel := make(chan error, 1)
	go func() {
		errChannel <- http.ListenAndServe(fmt.Sprintf(":%v", port), gqlHandler)
	}()

	ctx, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()
	select {
	case status := <-errChannel:
		return fmt.Errorf("could not start server because %v", status.Error())
	case <-ctx.Done():
		logger.Infof("Graphql Server started @ (port=%v)", port)
	}
	return nil
}
