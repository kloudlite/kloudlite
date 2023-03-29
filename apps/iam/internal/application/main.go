package application

import (
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/iam/internal/domain/entities"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/repos"
)

func fxServer(rbRepo repos.DbRepo[*entities.RoleBinding]) iam.IAMServer {
	return &GrpcServer{
		rbRepo: rbRepo,
	}
}

var Module = fx.Module(
	"application",
	fx.Provide(fxServer),
	repos.NewFxMongoRepo[*entities.RoleBinding]("role_bindings", "rb", entities.RoleBindingIndices),
	fx.Invoke(
		func(server *grpc.Server, iamService iam.IAMServer) {
			iam.RegisterIAMServer(server, iamService)
		},
	),
)
