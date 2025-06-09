package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type Env struct {
	SupportEmail   string `env:"SUPPORT_EMAIL" required:"true"`
	SendgridApiKey string `env:"SENDGRID_API_KEY" required:"true"`
	GrpcPort       uint16 `env:"GRPC_PORT" required:"true"`

	AccountsWebInviteUrl   string `env:"ACCOUNTS_WEB_INVITE_URL" required:"true"`
	ProjectsWebInviteUrl   string `env:"PROJECTS_WEB_INVITE_URL" required:"true"`
	KloudliteConsoleWebUrl string `env:"KLOUDLITE_CONSOLE_WEB_URL" required:"true"`

	ResetPasswordWebUrl string `env:"RESET_PASSWORD_WEB_URL" required:"true"`
	VerifyEmailWebUrl   string `env:"VERIFY_EMAIL_WEB_URL" required:"true"`

	AccountCookieName string `env:"ACCOUNT_COOKIE_NAME" required:"true"`

	NatsURL string `env:"NATS_URL" required:"true"`

	// NATS:start
	NotificationNatsStream string `env:"NOTIFICATION_NATS_STREAM" required:"true"`
	// NATS:start

	IsDev bool

	SessionKVBucket string `env:"SESSION_KV_BUCKET" required:"true"`
	IAMGrpcAddr     string `env:"IAM_GRPC_ADDR" required:"true"`

	CommsDBUri  string `env:"MONGO_URI" required:"true"`
	CommsDBName string `env:"MONGO_DB_NAME" required:"true"`

	Port uint16 `env:"HTTP_PORT" required:"true"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, errors.NewE(err)
	}
	return &ev, nil
}
