package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type Env struct {
	Port              uint16 `env:"PORT" required:"true"`
	GrpcPort          uint16 `env:"GRPC_PORT" required:"true"`
	CookieDomain      string `env:"COOKIE_DOMAIN" required:"true"`
	AccountCookieName string `env:"ACCOUNT_COOKIE_NAME" required:"true"`

	DBUri        string `env:"DB_URI" required:"true"`
	DBName       string `env:"DB_NAME" required:"true"`
	IAMGrpcAddr  string `env:"IAM_GRPC_ADDR" required:"true"`
	AuthGrpcAddr string `env:"AUTH_GRPC_ADDR" required:"true"`

	GithubClientId     string `env:"GITHUB_CLIENT_ID" required:"true"`
	GithubClientSecret string `env:"GITHUB_CLIENT_SECRET" required:"true"`
	GithubCallbackUrl  string `env:"GITHUB_CALLBACK_URL" required:"true"`
	GithubAppId        string `env:"GITHUB_APP_ID" required:"true"`
	GithubAppPKFile    string `env:"GITHUB_APP_PK_FILE" required:"true"`

	GithubScopes string `env:"GITHUB_SCOPES" required:"true"`

	// NATS:start
	NatsURL            string `env:"NATS_URL" required:"true"`
	ResourceNatsStream string `env:"RESOURCE_NATS_STREAM" required:"true"`
	EventsNatsStream   string `env:"EVENTS_NATS_STREAM" required:"true"`
	// NATS:end

	GitlabWebhookAuthzSecret string `env:"GITLAB_WEBHOOK_AUTHZ_SECRET" required:"true"`

	GitlabClientId     string `env:"GITLAB_CLIENT_ID" required:"true"`
	GitlabClientSecret string `env:"GITLAB_CLIENT_SECRET" required:"true"`
	GitlabCallbackUrl  string `env:"GITLAB_CALLBACK_URL" required:"true"`
	GitlabScopes       string `env:"GITLAB_SCOPES" required:"true"`
	GitlabWebhookUrl   string `env:"GITLAB_WEBHOOK_URL" required:"true"`

	RegistryHost           string `env:"REGISTRY_HOST" required:"true"`
	RegistrySecretKey      string `env:"REGISTRY_SECRET_KEY" required:"true"`
	RegistryAuthorizerPort uint16 `env:"REGISTRY_AUTHORIZER_PORT" required:"true"`

	JobBuildNamespace string `env:"JOB_BUILD_NAMESPACE" required:"true"`
	SessionKVBucket   string `env:"SESSION_KV_BUCKET" required:"true"`
	IsDev             bool
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, errors.NewE(err)
	}
	return &e, nil
}
