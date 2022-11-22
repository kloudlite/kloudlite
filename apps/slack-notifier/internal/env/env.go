package env

type Env struct {
	SlackAppToken  string `env:"SLACK_APP_TOKEN" required:"true"`
	SlackChannelID string `env:"SLACK_CHANNEL_ID" required:"true"`
	HttpPort       uint16 `env:"HTTP_PORT" required:"true"`
	HttpCors       string `env:"HTTP_CORS"`
}

type DevMode bool

func (m DevMode) Value() bool {
	return bool(m)
}
