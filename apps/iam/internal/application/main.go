package application

import (
	"encoding/json"
	"os"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/iam/internal/domain/entities"
	"kloudlite.io/apps/iam/internal/env"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
)

func fxServer(rbRepo repos.DbRepo[*entities.RoleBinding], rbm RoleBindingMap, logger logging.Logger) iam.IAMServer {
	return &GrpcServer{
		logger:         logger,
		rbRepo:         rbRepo,
		roleBindingMap: rbm,
	}
}

var Module = fx.Module(
	"application",
	fx.Provide(func(ev *env.Env) (RoleBindingMap, error) {
		if ev.ActionRoleMapFile != "" {
			b, err := os.ReadFile(ev.ActionRoleMapFile)
			if err != nil {
				return nil, err
			}
			var rbm RoleBindingMap
			if err := json.Unmarshal(b, &rbm); err != nil {
				return nil, err
			}
			return rbm, nil
		}

		return roleBindings, nil
	}),
	fx.Provide(fxServer),
	repos.NewFxMongoRepo[*entities.RoleBinding]("role_bindings", "rb", entities.RoleBindingIndices),
	fx.Invoke(
		func(server *grpc.Server, iamService iam.IAMServer) {
			iam.RegisterIAMServer(server, iamService)
		},
	),
)
