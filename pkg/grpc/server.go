package rpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"kloudlite.io/pkg/errors"
	"net"
)

func GRPCStartServer(_ context.Context, server *grpc.Server, port int) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return errors.NewEf(err, "could not listen to net/tcp server")
	}
	go func() error {
		err := server.Serve(listen)
		if err != nil {
			return errors.NewEf(err, "could not start grpc server ")
		}
		return nil
	}()
	return nil
}
