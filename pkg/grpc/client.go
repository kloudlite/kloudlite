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
	GetGRPCServerURL() string
}

type GrpcClient interface {
	~*grpc.ClientConn
}

func NewGrpcClientFx[T ClientOptions, M GrpcClient]() fx.Option {
	return fx.Module(
		"grpc-client",
		fx.Provide(
			func(env T) (M, error) {
				return NewInsecureClient(env.GetGRPCServerURL())
			},
		),
		fx.Invoke(
			func(grpcClient M, lifecycle fx.Lifecycle) {
				lifecycle.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							return nil
						},
						OnStop: func(ctx context.Context) error {
							return nil
						},
					},
				)
			},
		),
	)
}
