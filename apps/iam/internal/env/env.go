package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type Env struct {
	GrpcPort    uint16 `env:"GRPC_PORT" required:"true"`
	MongoDbUri  string `env:"MONGO_DB_URI" required:"true"`
	MongoDbName string `env:"MONGO_DB_NAME" required:"true"`

	ActionRoleMapFile string `env:"ACTION_ROLE_MAP_FILE" required:"false"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, errors.NewE(err)
	}
	return &e, nil
}
