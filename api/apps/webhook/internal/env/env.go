package env

type Env struct {
	HttpPort uint16 `env:"HTTP_PORT" required:"true"`

	KlHookTriggerAuthzSecret string `env:"KL_HOOK_TRIGGER_AUTHZ_SECRET" required:"true"`

	GitWebhooksTopic  string `env:"GIT_WEBHOOKS_TOPIC" required:"false"`
	GithubAuthzSecret string `env:"GITHUB_AUTHZ_SECRET" required:"false"`
	GitlabAuthzSecret string `env:"GITLAB_AUTHZ_SECRET" required:"false"`
	NatsURL           string `env:"NATS_URL" required:"false"`

	CommsService      string `env:"COMMS_SERVICE" required:"true"`
	DiscordWebhookUrl string `env:"DISCORD_WEBHOOK_URL" required:"false"`
	WebhookURL        string `env:"WEBHOOK_URL" required:"true"`

	WebhookTokenHashingSecret string `env:"WEBHOOK_TOKEN_HASHING_SECRET" required:"true"`
}
