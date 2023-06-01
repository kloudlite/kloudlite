package domain

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/nodectrl/internal/env"
	mongogridfs "kloudlite.io/pkg/mongo-gridfs"
)

type domain struct {
	env *env.Env
	gfs mongogridfs.GridFs
}

var Module = fx.Module("domain",
	fx.Provide(
		func(env *env.Env, gfs mongogridfs.GridFs) Domain {
			return domain{
				env: env,
				gfs: gfs,
			}
		},
	),
	ProviderClientFx,
)
