package env

type Env struct {
	CiAddr   string `env:"CI_ADDR" required:"true"`
	HttpPort uint16 `env:"HTTP_PORT" required:"true"`

	KafkaBrokers     string `env:"KAFKA_BROKERS" required:"true"`
	GitWebhooksTopic string `env:"GIT_WEBHOOKS_TOPIC" required:"true"`

	HarborWebhookTopic string `env:"HARBOR_WEBHOOK_TOPIC" required:"true"`
	HarborAuthzSecret  string `env:"HARBOR_AUTHZ_SECRET" required:"true"`
}
