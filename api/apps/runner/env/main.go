package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type mainEnv struct {
	MongoUri    string `env:"MONGO_URI" required:"true"`
	MongoDbName string `env:"MONGO_DB_NAME" required:"true"`
	NatsURL     string `env:"NATS_URL" required:"true"`
	GrpcPort    uint16 `env:"GRPC_PORT" required:"true"`

	CommsServiceAddr    string `env:"COMMS_SERVICE" required:"true"`
	ConsoleServiceAddr  string `env:"CONSOLE_SERVICE" required:"true"`
	AuthServiceAddr     string `env:"AUTH_SERVICE" required:"true"`
	InfraServiceAddr    string `env:"INFRA_SERVICE" required:"true"`
	IAMServiceAddr      string `env:"IAM_SERVICE" required:"true"`
	AccountsServiceAddr string `env:"ACCOUNT_SERVICE" required:"true"`

	IsDev bool
}

type Env struct {
	mainEnv
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev.mainEnv); err != nil {
		return nil, errors.NewE(err)
	}
	return &ev, nil
}
