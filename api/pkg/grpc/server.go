package rpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"net"
)

func GRPCStartServer(ctx context.Context, server *grpc.Server, port int) error {
	listen, portFormatError := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if portFormatError != nil {
		return portFormatError
	}
	go func() error {
		serverStartError := server.Serve(listen)
		if serverStartError != nil {
			return serverStartError
		}
		return nil
	}()
	return nil
}
