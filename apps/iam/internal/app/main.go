package app

import (
	"encoding/json"
	"github.com/kloudlite/api/apps/iam/internal/entities"
	"github.com/kloudlite/api/pkg/logging"
	"os"

	"github.com/kloudlite/api/apps/iam/internal/env"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
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
