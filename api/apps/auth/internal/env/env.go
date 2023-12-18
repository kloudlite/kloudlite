package env

import (
	"fmt"

	"github.com/codingconcepts/env"
)

type Env struct {
	MongoUri      string `env:"MONGO_URI" required:"true"`
	MongoDbName   string `env:"MONGO_DB_NAME" required:"true"`
	Port          uint16 `env:"PORT" required:"true"`
	GrpcPort      uint16 `env:"GRPC_PORT" required:"true"`
	CorsOrigins   string `env:"ORIGINS"`

	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	OAuth2Enabled bool `env:"OAUTH2_ENABLED" required:"true"`

	OAuth2GithubEnabled bool   `env:"OAUTH2_GITHUB_ENABLED" required:"false"`
	GithubClientId      string `env:"GITHUB_CLIENT_ID" required:"false"`
	GithubClientSecret  string `env:"GITHUB_CLIENT_SECRET" required:"false"`
	GithubCallbackUrl   string `env:"GITHUB_CALLBACK_URL" required:"false"`
	GithubAppId         string `env:"GITHUB_APP_ID" required:"false"`
	GithubAppPKFile     string `env:"GITHUB_APP_PK_FILE" required:"false"`
	GithubScopes        string `env:"GITHUB_SCOPES" required:"false"`
	GithubWebhookUrl    string `env:"GITHUB_WEBHOOK_URL" required:"false"`

	OAuth2GitlabEnabled bool   `env:"OAUTH2_GITLAB_ENABLED" required:"false"`
	GitlabClientId      string `env:"GITLAB_CLIENT_ID" required:"false"`
	GitlabClientSecret  string `env:"GITLAB_CLIENT_SECRET" required:"false"`
	GitlabCallbackUrl   string `env:"GITLAB_CALLBACK_URL" required:"false"`
	GitlabScopes        string `env:"GITLAB_SCOPES" required:"false"`
	GitlabWebhookUrl    string `env:"GITLAB_WEBHOOK_URL" required:"false"`

	OAuth2GoogleEnabled bool   `env:"OAUTH2_GOOGLE_ENABLED" required:"false"`
	GoogleClientId      string `env:"GOOGLE_CLIENT_ID" required:"false"`
	GoogleClientSecret  string `env:"GOOGLE_CLIENT_SECRET" required:"false"`
	GoogleCallbackUrl   string `env:"GOOGLE_CALLBACK_URL" required:"false"`
	GoogleScopes        string `env:"GOOGLE_SCOPES" required:"false"`

	CommsService    string `env:"COMMS_SERVICE" required:"true"`
	NatsURL         string `env:"NATS_URL" required:"true"`
	SessionKVBucket string `env:"SESSION_KV_BUCKET" required:"true"`
}

func (ev *Env) validateEnv() error {
	if ev.OAuth2Enabled {
		if ev.OAuth2GithubEnabled {
			err := fmt.Errorf("when github oauth2 is enabled, secrets `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_CALLBACK_URL`, `GITHUB_APP_ID`, `GITHUB_APP_PK_FILE`, `GITHUB_SCOPES` are required")

			if ev.GithubClientId == "" ||
				ev.GithubClientSecret == "" ||
				ev.GithubCallbackUrl == "" ||
				ev.GithubAppId == "" ||
				ev.GithubAppPKFile == "" ||
				ev.GithubScopes == "" {
				return err
			}
		}

		if ev.OAuth2GitlabEnabled {
			err := fmt.Errorf("when gitlab oauth2 is enabled, secrets `GITLAB_CLIENT_ID`, `GITLAB_CLIENT_SECRET`, `GITLAB_CALLBACK_URL`, `GITLAB_SCOPES` are required")

			if ev.GitlabClientId == "" ||
				ev.GitlabClientSecret == "" ||
				ev.GitlabCallbackUrl == "" ||
				ev.GitlabScopes == "" {
				return err
			}
		}

		if ev.OAuth2GoogleEnabled {
			err := fmt.Errorf("when google oauth2 is enabled, secrets `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_CALLBACK_URL`, `GOOGLE_SCOPES` are required")

			if ev.GoogleClientId == "" ||
				ev.GoogleClientSecret == "" ||
				ev.GoogleCallbackUrl == "" ||
				ev.GoogleScopes == "" {
				return err
			}
		}
	}
	return nil
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	if err := ev.validateEnv(); err != nil {
		return nil, err
	}
	return &ev, nil
}
