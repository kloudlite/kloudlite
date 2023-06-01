package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/nodectrl/internal/app"
	"kloudlite.io/apps/nodectrl/internal/env"
	mongogridfs "kloudlite.io/pkg/mongo-gridfs"
)

type fm struct {
	env *env.Env
}

func (fm *fm) GetMongoConfig() (url string, dbName string) {
	return fm.env.DBUrl, fm.env.DBName
}

var Module = fx.Module(
	"framework",
	fx.Provide(func(env *env.Env) *fm {
		return &fm{env}
	}),
	mongogridfs.NewMongoGridFsClientFx[*fm](),
	app.Module,
)
