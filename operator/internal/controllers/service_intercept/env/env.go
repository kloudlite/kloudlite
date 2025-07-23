package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES" default:"5"`

	KloudliteNamespace string `env:"KLOUDLITE_NAMESPACE" default:"kloudlite"`

	ServiceInterceptWebhookServiceSelector map[string]string `env:"SERVICE_INTERCEPT_WEBHOOK_SERVICE_SELECTOR" default:"app=service-intercept"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}

