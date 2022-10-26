package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/app"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	loki_server "kloudlite.io/pkg/loki-server"
	"kloudlite.io/pkg/redpanda"
	mongo_db "kloudlite.io/pkg/repos"
	rcn "kloudlite.io/pkg/res-change-notifier"
)

type GrpcDNSConfig struct {
	DNSService string `env:"DNS_SERVICE" required:"true"`
}

func (e *GrpcDNSConfig) GetGRPCServerURL() string {
	return e.DNSService
}

type GrpcFinanceConfig struct {
	FinanceService string `env:"FINANCE_SERVICE" required:"true"`
}

func (e *GrpcFinanceConfig) GetGRPCServerURL() string {
	return e.FinanceService
}

type GrpcAuthConfig struct {
	AuthService string `env:"AUTH_SERVICE" required:"true"`
}

func (e *GrpcAuthConfig) GetGRPCServerURL() string {
	return e.AuthService
}

type GrpcCIConfig struct {
	CIService string `env:"CI_SERVICE" required:"true"`
}

func (e *GrpcCIConfig) GetGRPCServerURL() string {
	return e.CIService
}

type IAMGRPCEnv struct {
	IAMService string `env:"IAM_SERVICE"`
}

func (e *IAMGRPCEnv) GetGRPCServerURL() string {
	return e.IAMService
}

type JSEvalEnv struct {
	JSEvalService string `env:"JSEVAL_SERVICE"`
}

func (e *JSEvalEnv) GetGRPCServerURL() string {
	return e.JSEvalService
}

type LogServerEnv struct {
	LokiServerUrl    string `env:"LOKI_URL" required:"true"`
	LokiAuthUsername string `env:"LOKI_AUTH_USERNAME"`
	LokiAuthPassword string `env:"LOKI_AUTH_PASSWORD"`
	LogServerPort    uint64 `env:"LOG_SERVER_PORT" required:"true"`
}

func (l *LogServerEnv) GetLokiServerUrlAndOptions() (string, loki_server.ClientOpts) {
	opts := loki_server.ClientOpts{}
	if l.LokiAuthUsername != "" && l.LokiAuthPassword != "" {
		opts.BasicAuth = &loki_server.BasicAuth{
			Username: l.LokiAuthUsername,
			Password: l.LokiAuthPassword,
		}
	}
	return l.LokiServerUrl, opts
}

func (l *LogServerEnv) GetLogServerPort() uint64 {
	return l.LogServerPort
}

type Env struct {
	MongoUri      string `env:"MONGO_URI" required:"true"`
	RedisHosts    string `env:"REDIS_HOSTS" required:"true"`
	RedisUserName string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisPrefix   string `env:"REDIS_PREFIX"`

	AuthRedisHosts    string `env:"REDIS_AUTH_HOSTS" required:"true"`
	AuthRedisUserName string `env:"REDIS_AUTH_USERNAME"`
	AuthRedisPassword string `env:"REDIS_AUTH_PASSWORD"`
	AuthRedisPrefix   string `env:"REDIS_AUTH_PREFIX" required:"true"`

	MongoDbName   string `env:"MONGO_DB_NAME" required:"true"`
	KafkaBrokers  string `env:"KAFKA_BOOTSTRAP_SERVERS" required:"true"`
	KafkaUsername string `env:"KAFKA_USERNAME" required:"true"`
	KafkaPassword string `env:"KAFKA_PASSWORD" required:"true"`

	Port  uint16 `env:"PORT" required:"true"`
	IsDev bool   `env:"DEV" default:"false" required:"true"`

	GrpcPort    uint16 `env:"GRPC_PORT" required:"true"`
	NotifierUrl string `env:"NOTIFIER_URL" required:"true"`

	KubeAPIAddress string `env:"KUBE_API_ADDRESS" required:"true"`
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

func (e *Env) GetHttpPort() uint16 {
	return e.Port
}

func (e *Env) GetHttpCors() string {
	return ""
}

func (e *Env) RedisOptions() (hosts, username, password, basePrefix string) {
	return e.RedisHosts, e.RedisUserName, e.RedisPassword, e.RedisPrefix
}

func (e *Env) GetMongoConfig() (url string, dbName string) {
	return e.MongoUri, e.MongoDbName
}

func (e *Env) GetGRPCPort() uint16 {
	return e.GrpcPort
}

func (e *Env) GetNotifierUrl() string {
	return e.NotifierUrl
}

var Module fx.Option = fx.Module(
	"framework",
	config.EnvFx[Env](),
	config.EnvFx[LogServerEnv](),
	config.EnvFx[IAMGRPCEnv](),
	config.EnvFx[JSEvalEnv](),

	config.EnvFx[GrpcAuthConfig](),
	config.EnvFx[GrpcDNSConfig](),
	config.EnvFx[GrpcFinanceConfig](),
	config.EnvFx[GrpcCIConfig](),

	redpanda.NewClientFx[*Env](),
	rcn.NewFxResourceChangeNotifier[*Env](),
	rpc.NewGrpcServerFx[*Env](),
	rpc.NewGrpcClientFx[*IAMGRPCEnv, app.IAMClientConnection](),
	rpc.NewGrpcClientFx[*JSEvalEnv, app.JSEvalClientConnection](),
	rpc.NewGrpcClientFx[*GrpcAuthConfig, app.AuthClientConnection](),
	rpc.NewGrpcClientFx[*GrpcCIConfig, app.CIClientConnection](),
	rpc.NewGrpcClientFx[*GrpcFinanceConfig, app.FinanceClientConnection](),
	rpc.NewGrpcClientFx[*GrpcDNSConfig, app.DNSClientConnection](),

	mongo_db.NewMongoClientFx[*Env](),
	fx.Provide(
		func(env *Env) app.AuthCacheClient {
			return cache.NewRedisClient(env.AuthRedisHosts, env.AuthRedisUserName, env.AuthRedisPassword, env.AuthRedisPrefix)
		},
	),
	cache.FxLifeCycle[app.AuthCacheClient](),

	fx.Provide(
		func(env *Env) app.CacheClient {
			return cache.NewRedisClient(env.RedisOptions())
		},
	),
	cache.FxLifeCycle[app.CacheClient](),
	httpServer.NewHttpServerFx[*Env](),
	loki_server.NewLogServerFx[*LogServerEnv](), // will provide log server and loki client
	app.Module,
)
