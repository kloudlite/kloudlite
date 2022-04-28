package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/app"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
	mongo_db "kloudlite.io/pkg/repos"
)

type GrpcInfraConfig struct {
	InfraGrpcHost string `env:"INFRA_HOST" required:"true"`
	InfraGrpcPort string `env:"INFRA_PORT" required:"true"`
}

func (e *GrpcInfraConfig) GetGCPServerURL() string {
	return e.InfraGrpcHost + ":" + e.InfraGrpcPort
}

type GrpcAuthConfig struct {
	InfraGrpcHost string `env:"AUTH_HOST" required:"true"`
	InfraGrpcPort string `env:"AUTH_PORT" required:"true"`
}

func (e *GrpcAuthConfig) GetGCPServerURL() string {
	return e.InfraGrpcHost + ":" + e.InfraGrpcPort
}

type GrpcCIConfig struct {
	CIGrpcHost string `env:"CI_HOST" required:"true"`
	CIGrpcPort string `env:"CI_PORT" required:"true"`
}

func (e *GrpcCIConfig) GetGCPServerURL() string {
	return e.CIGrpcHost + ":" + e.CIGrpcPort
}

type Env struct {
	MongoUri      string `env:"MONGO_URI" required:"true"`
	RedisHosts    string `env:"REDIS_HOSTS" required:"true"`
	RedisUserName string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	MongoDbName   string `env:"MONGO_DB_NAME" required:"true"`
	KafkaBrokers  string `env:"KAFKA_BOOTSTRAP_SERVERS" required:"true"`
	Port          uint16 `env:"PORT" required:"true"`
	IsDev         bool   `env:"DEV" default:"false"`
	CorsOrigins   string `env:"ORIGINS"`
	GrpcPort      uint16 `env:"GRPC_PORT"`
}

func (e *Env) GetBrokers() string {
	return e.KafkaBrokers
}

func (e *Env) GetHttpPort() uint16 {
	return e.Port
}

func (e *Env) GetHttpCors() string {
	return e.CorsOrigins
}

func (e *Env) RedisOptions() (hosts, username, password string) {
	return e.RedisHosts, e.RedisUserName, e.RedisPassword
}

func (e *Env) GetMongoConfig() (url string, dbName string) {
	return e.MongoUri, e.MongoDbName
}

func (e *Env) GetGRPCPort() uint16 {
	return e.GrpcPort
}

var Module = fx.Module("framework",
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(config.LoadEnv[GrpcInfraConfig]()),
	fx.Provide(config.LoadEnv[GrpcAuthConfig]()),
	fx.Provide(config.LoadEnv[GrpcCIConfig]()),
	fx.Provide(logger.NewLogger),
	rpc.NewGrpcServerFx[*Env](),
	rpc.NewGrpcClientFx[*GrpcInfraConfig, app.InfraClientConnection](),
	rpc.NewGrpcClientFx[*GrpcAuthConfig, app.AuthClientConnection](),
	rpc.NewGrpcClientFx[*GrpcCIConfig, app.CIClientConnection](),
	mongo_db.NewMongoClientFx[*Env](),
	messaging.NewKafkaClientFx[*Env](),
	cache.NewRedisFx[*Env](),
	httpServer.NewHttpServerFx[*Env](),

	app.Module,
)
