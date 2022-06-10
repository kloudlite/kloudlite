package framework

import (
	"go.uber.org/fx"
	httpServer "kloudlite.io/pkg/http-server"
)

type Env struct {
	HttpPort uint16 `env:"HTTP_PORT" required:"true"`
	HttpCors string `env:"HTTP_CORS_ORIGINS" required:"true"`
}

func (e Env) GetHttpPort() uint16 {
	return e.HttpPort
}

func (e Env) GetHttpCors() string {
	return e.HttpCors
}

var Module = fx.Module(
	"tekton-interceptor",
	httpServer.NewHttpServerFx[*Env](),
)
