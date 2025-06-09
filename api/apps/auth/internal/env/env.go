package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type authEnv struct {
	MongoUri    string `env:"MONGO_URI" required:"true"`
	MongoDbName string `env:"MONGO_DB_NAME" required:"true"`
	Port        uint16 `env:"PORT" required:"true"`
	GrpcPort    uint16 `env:"GRPC_PORT" required:"true"`
	CorsOrigins string `env:"ORIGINS"`

	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	UserEmailVerifactionEnabled bool `env:"USER_EMAIL_VERIFICATION_ENABLED" default:"true"`

	OAuth2Enabled bool `env:"OAUTH2_ENABLED" required:"true"`

	OAuth2GithubEnabled bool `env:"OAUTH2_GITHUB_ENABLED" default:"false"`
	OAuth2GitlabEnabled bool `env:"OAUTH2_GITLAB_ENABLED" default:"false"`
	OAuth2GoogleEnabled bool `env:"OAUTH2_GOOGLE_ENABLED" default:"false"`

	CommsService               string `env:"COMMS_SERVICE" required:"true"`
	NatsURL                    string `env:"NATS_URL" required:"true"`
	SessionKVBucket            string `env:"SESSION_KV_BUCKET" required:"true"`
	VerifyTokenKVBucket        string `env:"VERIFY_TOKEN_KV_BUCKET" required:"true"`
	ResetPasswordTokenKVBucket string `env:"RESET_PASSWORD_TOKEN_KV_BUCKET" required:"true"`

	GoogleRecaptchaEnabled bool `env:"GOOGLE_RECAPTCHA_ENABLED" default:"false"`

	IsDev bool
}

type githubOAuthEnv struct {
	GithubClientId     string `env:"GITHUB_CLIENT_ID" required:"true"`
	GithubClientSecret string `env:"GITHUB_CLIENT_SECRET" required:"true"`
	GithubCallbackUrl  string `env:"GITHUB_CALLBACK_URL" required:"true"`
	GithubAppId        string `env:"GITHUB_APP_ID" required:"true"`
	GithubAppPKFile    string `env:"GITHUB_APP_PK_FILE" required:"true"`
	GithubScopes       string `env:"GITHUB_SCOPES" required:"true"`
	GithubWebhookUrl   string `env:"GITHUB_WEBHOOK_URL" required:"false"`
}

type gitlabOAuthEnv struct {
	GitlabClientId     string `env:"GITLAB_CLIENT_ID" required:"true"`
	GitlabClientSecret string `env:"GITLAB_CLIENT_SECRET" required:"true"`
	GitlabCallbackUrl  string `env:"GITLAB_CALLBACK_URL" required:"true"`
	GitlabScopes       string `env:"GITLAB_SCOPES" required:"true"`
}

type googleOAuthEnv struct {
	GoogleClientId     string `env:"GOOGLE_CLIENT_ID" required:"true"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET" required:"true"`
	GoogleCallbackUrl  string `env:"GOOGLE_CALLBACK_URL" required:"true"`
	GoogleScopes       string `env:"GOOGLE_SCOPES" required:"true"`
}

type googleRecaptchaEnv struct {
	GoogleCloudProjectId         string `env:"GOOGLE_CLOUD_PROJECT_ID" required:"true"`
	RecaptchaSiteKey             string `env:"RECAPTCHA_SITE_KEY" required:"true"`
	GoogleApplicationCredentials string `env:"GOOGLE_APPLICATION_CREDENTIALS" required:"true"`
}

type Env struct {
	authEnv
	githubOAuthEnv
	gitlabOAuthEnv
	googleOAuthEnv
	googleRecaptchaEnv
}

func LoadEnv() (*Env, error) {
	var ev Env

	if err := env.Set(&ev.authEnv); err != nil {
		return nil, errors.NewE(err)
	}

	if ev.OAuth2GithubEnabled {
		if err := env.Set(&ev.githubOAuthEnv); err != nil {
			return nil, errors.NewE(err)
		}
	}

	if ev.OAuth2GitlabEnabled {
		if err := env.Set(&ev.gitlabOAuthEnv); err != nil {
			return nil, errors.NewE(err)
		}
	}

	if ev.OAuth2GoogleEnabled {
		if err := env.Set(&ev.googleOAuthEnv); err != nil {
			return nil, errors.NewE(err)
		}
	}

	if ev.GoogleRecaptchaEnabled {
		if err := env.Set(&ev.googleRecaptchaEnv); err != nil {
			return nil, errors.NewE(err)
		}
	}

	return &ev, nil
}
