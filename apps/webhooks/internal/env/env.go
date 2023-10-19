package env

type Env struct {
	HttpPort uint16 `env:"HTTP_PORT" required:"true"`

	KafkaBrokers  string `env:"KAFKA_BROKERS" required:"true"`

	KlHookTriggerAuthzSecret string `env:"KL_HOOK_TRIGGER_AUTHZ_SECRET" required:"true"`

	GitWebhooksTopic  string `env:"GIT_WEBHOOKS_TOPIC" required:"false"`
	GithubAuthzSecret string `env:"GITHUB_AUTHZ_SECRET" required:"false"`
	GitlabAuthzSecret string `env:"GITLAB_AUTHZ_SECRET" required:"false"`
}
