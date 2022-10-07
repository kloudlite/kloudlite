package env

type Env struct {
	CiAddr   string `env:"CI_ADDR" required:"true"`
	HttpPort uint16 `env:"HTTP_PORT" required:"true"`

	KafkaUsername    string `env:"KAFKA_USERNAME" required:"true"`
	KafkaPassword    string `env:"KAFKA_PASSWORD" required:"true"`
	KafkaBrokers             string `env:"KAFKA_BROKERS" required:"true"`
	GitWebhooksTopic         string `env:"GIT_WEBHOOKS_TOPIC" required:"true"`
	KlHookTriggerAuthzSecret string `env:"KL_HOOK_TRIGGER_AUTHZ_SECRET" required:"true"`

	GithubAuthzSecret string `env:"GITHUB_AUTHZ_SECRET" required:"true"`
	GitlabAuthzSecret string `env:"GITLAB_AUTHZ_SECRET" required:"true"`

	HarborWebhookTopic string `env:"HARBOR_WEBHOOK_TOPIC" required:"true"`
	HarborAuthzSecret  string `env:"HARBOR_AUTHZ_SECRET" required:"true"`
}
