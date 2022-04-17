package rpc

import (
	"context"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewInsecureClient(grpcUrl string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type ClientOptions interface {
	GetGCPServerURL() string
}

func NewGrpcClientFx[T ClientOptions]() fx.Option {
	return fx.Module(
		"grpc-client",
		fx.Provide(func(env T) (*grpc.ClientConn, error) {
			return NewInsecureClient(env.GetGCPServerURL())
		}),
		fx.Invoke(func(grpcClient *grpc.ClientConn, lifecycle fx.Lifecycle) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return grpcClient.Close()
				},
			})
		}),
	)
}
