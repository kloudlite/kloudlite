package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type authEnv struct {
	UserEmailVerificationEnabled bool   `env:"AUTH__USER_EMAIL_VERIFICATION_ENABLED" default:"true"`
	VerifyTokenKVBucket          string `env:"AUTH__VERIFY_TOKEN_KV_BUCKET" required:"true"`
	ResetPasswordTokenKVBucket   string `env:"AUTH__RESET_PASSWORD_TOKEN_KV_BUCKET" required:"true"`
	
	// JWT Configuration
	JWTSecret                    string `env:"AUTH__JWT_SECRET" required:"true"`
	JWTTokenExpiry               string `env:"AUTH__JWT_TOKEN_EXPIRY" default:"15m"`
	JWTRefreshTokenExpiry        string `env:"AUTH__JWT_REFRESH_TOKEN_EXPIRY" default:"168h"`
	
	// Web Configuration
	WebUrl                       string `env:"AUTH__WEB_URL" default:"http://localhost:3000"`
	
	// Email Configuration
	SupportEmail                 string `env:"AUTH__SUPPORT_EMAIL" required:"true"`
	MailtrapApiToken             string `env:"AUTH__MAILTRAP_API_TOKEN" required:"true"`
	
	// Platform Configuration
	PlatformOwnerEmail           string `env:"AUTH__PLATFORM_OWNER_EMAIL"`
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
