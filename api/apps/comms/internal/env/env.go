package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type CommsEnv struct {
	BaseUrl                string `env:"COMMS__BASE_URL" required:"true"`
	SupportEmail           string `env:"COMMS__SUPPORT_EMAIL" required:"true"`
	SendgridApiKey         string `env:"COMMS__SENDGRID_API_KEY" required:"true"`
	NotificationNatsStream string `env:"COMMS__NOTIFICATION_NATS_STREAM" required:"true"`
}

func LoadEnv() (*CommsEnv, error) {
	var ev CommsEnv
	if err := env.Set(&ev); err != nil {
		return nil, errors.NewE(err)
	}
	return &ev, nil
}
