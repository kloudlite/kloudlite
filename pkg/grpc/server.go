package rpc

import (
	"context"
	"fmt"
	"go.uber.org/fx"
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

type ServerOptions interface {
	GetGRPCPort() uint16
}

func NewGrpcServerFx[T ServerOptions]() fx.Option {
	return fx.Module("grpc-server",
		fx.Provide(grpc.NewServer),
		fx.Invoke(func(lf fx.Lifecycle, env T, server *grpc.Server) {
			lf.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					listen, err := net.Listen("tcp", fmt.Sprintf(":%d", env.GetGRPCPort()))
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
				},
				OnStop: func(context.Context) error {
					server.Stop()
					return nil
				},
			})
		}),
	)
}
