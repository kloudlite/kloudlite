package env

import (
	"github.com/codingconcepts/env"
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
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
