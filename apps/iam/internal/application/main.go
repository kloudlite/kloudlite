package application

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
)

type IamServerI struct {
	iam.UnimplementedIAMServiceServer
}

func (i *IamServerI) Ping(ctx context.Context, in *iam.Message) (*iam.Message, error) {
	fmt.Println("Ping received", in.Message)
	return &iam.Message{
		Message: "pong",
	}, nil
}

func fxIamService() iam.IAMServiceServer {
	return &IamServerI{}
}

var Module = fx.Module("application",
	fx.Provide(fxIamService),
	fx.Invoke(func(server *grpc.Server, iamService iam.IAMServiceServer) {
		iam.RegisterIAMServiceServer(server, iamService)
	}),
)
