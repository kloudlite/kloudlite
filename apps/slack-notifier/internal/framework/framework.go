package framework

import (
	"github.com/slack-go/slack"
	"go.uber.org/fx"
	"kloudlite.io/apps/slack-notifier/internal/app"
	"kloudlite.io/apps/slack-notifier/internal/env"
	httpServer "kloudlite.io/pkg/http-server"
)

type fm struct {
	ev *env.Env
}

func (f fm) GetHttpPort() uint16 {
	return f.ev.HttpPort
}

func (f fm) GetHttpCors() string {
	return f.ev.HttpCors
}

var Module = fx.Module(
	"framework",
	fx.Provide(
		func(ev *env.Env) *fm {
			return &fm{ev: ev}
		},
	),
	fx.Provide(
		func(ev *env.Env, devMode env.DevMode) *slack.Client {
			return slack.New(ev.SlackAppToken, slack.OptionDebug(devMode.Value()))
		},
	),
	httpServer.NewHttpServerFx[*fm](),
	app.Module,
)
