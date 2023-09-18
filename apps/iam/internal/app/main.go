package app

import (
	"encoding/json"
	"kloudlite.io/apps/iam/internal/entities"
	"kloudlite.io/pkg/logging"
	"os"

	"go.uber.org/fx"
	"kloudlite.io/apps/iam/internal/env"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/repos"
)

var Module = fx.Module(
	"app",
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

	repos.NewFxMongoRepo[*entities.RoleBinding]("role_bindings", "rb", entities.RoleBindingIndices),

	fx.Provide(func(logger logging.Logger, rbRepo repos.DbRepo[*entities.RoleBinding], rbm RoleBindingMap) iam.IAMServer {
		return newIAMGrpcService(logger, rbRepo, rbm)
	}),

	fx.Invoke(
		func(server IAMGrpcServer, iamService iam.IAMServer) {
			iam.RegisterIAMServer(server, iamService)
		},
	),
)
