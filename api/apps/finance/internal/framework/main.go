package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/finance/internal/app"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
)

type AuthGRPCEnv struct {
	AuthService string `env:"AUTH_SERVICE"`
}

func (e *AuthGRPCEnv) GetGRPCServerURL() string {
	return e.AuthService
}

type CommsGRPCEnv struct {
	CommsGrpcService string `env:"COMMS_SERVICE"`
}

func (e *CommsGRPCEnv) GetGRPCServerURL() string {
	return e.CommsGrpcService
}

type ConsoleGRPCEnv struct {
	ConsoleGrpcService string `env:"CONSOLE_SERVICE"`
}

func (e *ConsoleGRPCEnv) GetGRPCServerURL() string {
	return e.ConsoleGrpcService
}

type IAMGRPCEnv struct {
	IAMService string `env:"IAM_SERVICE"`
}

func (e *IAMGRPCEnv) GetGRPCServerURL() string {
	return e.IAMService
}

type Env struct {
	DBName        string `env:"MONGO_DB_NAME"`
	DBUrl         string `env:"MONGO_URI"`
	KafkaBrokers  string `env:"WORKLOAD_KAFKA_BROKERS" required:"true"`
	KafkaUsername string `env:"KAFKA_USERNAME" required:"true"`
	KafkaPassword string `env:"KAFKA_PASSWORD" required:"true"`

	RedisHosts    string `env:"REDIS_HOSTS"`
	RedisUsername string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisPrefix   string `env:"REDIS_PREFIX"`

	AuthRedisHosts    string `env:"REDIS_AUTH_HOSTS"`
	AuthRedisUserName string `env:"REDIS_AUTH_USERNAME"`
	AuthRedisPassword string `env:"REDIS_AUTH_PASSWORD"`
	AuthRedisPrefix   string `env:"REDIS_AUTH_PREFIX"`

	HttpPort uint16 `env:"PORT"`
	HttpCors string `env:"ORIGINS"`
}

func (e *Env) GetKafkaSASLAuth() *redpanda.KafkaSASLAuth {
	return &redpanda.KafkaSASLAuth{
		SASLMechanism: redpanda.ScramSHA256,
		User:          e.KafkaUsername,
		Password:      e.KafkaPassword,
	}
}

func (e *Env) GetBrokers() string {
	return e.KafkaBrokers
}

func (e *Env) GetMongoConfig() (url string, dbName string) {
	return e.DBUrl, e.DBName
}

func (e *Env) RedisOptions() (hosts, username, password, basePrefix string) {
	return e.RedisHosts, e.RedisUsername, e.RedisPassword, e.RedisPrefix
}

func (e *Env) GetHttpPort() uint16 {
	return e.HttpPort
}

func (e *Env) GetHttpCors() string {
	return e.HttpCors
}

var Module fx.Option = fx.Module(
	"framework",
	config.EnvFx[Env](),
	config.EnvFx[ConsoleGRPCEnv](),
	config.EnvFx[CommsGRPCEnv](),
	config.EnvFx[IAMGRPCEnv](),
	config.EnvFx[AuthGRPCEnv](),
	rpc.NewGrpcClientFx[*ConsoleGRPCEnv, app.ConsoleClientConnection](),
	rpc.NewGrpcClientFx[*CommsGRPCEnv, app.CommsClientConnection](),
	rpc.NewGrpcClientFx[*IAMGRPCEnv, app.IAMClientConnection](),
	rpc.NewGrpcClientFx[*AuthGRPCEnv, app.AuthGrpcClientConn](),
	repos.NewMongoClientFx[*Env](),
	redpanda.NewClientFx[*Env](),
	fx.Provide(
		func(env *Env) app.AuthCacheClient {
			return cache.NewRedisClient(
				env.AuthRedisHosts,
				env.AuthRedisUserName,
				env.AuthRedisPassword,
				env.AuthRedisPrefix,
			)
		},
	),
	cache.FxLifeCycle[app.AuthCacheClient](),

	fx.Provide(
		func(env *Env) cache.Client {
			return cache.NewRedisClient(env.RedisOptions())
		},
	),
	cache.FxLifeCycle[cache.Client](),
	httpServer.NewHttpServerFx[*Env](),
	app.Module,
)
