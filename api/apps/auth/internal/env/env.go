package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type authEnv struct {
	UserEmailVerificationEnabled bool   `env:"AUTH__USER_EMAIL_VERIFICATION_ENABLED" default:"true"`
	VerifyTokenKVBucket          string `env:"AUTH__VERIFY_TOKEN_KV_BUCKET" required:"true"`
	ResetPasswordTokenKVBucket   string `env:"AUTH__RESET_PASSWORD_TOKEN_KV_BUCKET" required:"true"`
}

type AuthEnv struct {
	authEnv
}

func LoadEnv() (*AuthEnv, error) {
	var ev AuthEnv
	if err := env.Set(&ev.authEnv); err != nil {
		return nil, errors.NewE(err)
	}
	return &ev, nil
}
