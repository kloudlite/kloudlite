package application

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
)

type iamServer struct {
	iam.UnimplementedIAMServer
}

func (iamserver *iamServer) Can(ctx context.Context, input *iam.CanInput) (*iam.CanResult, error) {
	return nil, nil
}

func (iamserver *iamServer) ListMemberships(ctx context.Context, input *iam.Membership) (*iam.RoleBindingResp, error) {
	return nil, nil
}

func (i *iamServer) Ping(ctx context.Context, in *iam.Message) (*iam.Message, error) {
	fmt.Println("Ping received", in.Message)
	return &iam.Message{
		Message: "pong",
	}, nil
}

func fxIamService() iam.IAMServer {
	return &iamServer{}
}

var Module = fx.Module("application",
	fx.Provide(fxIamService),
	fx.Invoke(func(server *grpc.Server, iamService iam.IAMServer) {
		iam.RegisterIAMServer(server, iamService)
	}),
)
