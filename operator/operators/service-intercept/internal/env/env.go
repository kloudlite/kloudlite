package env

import (
	"encoding/json"
	"os"

	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles                int    `env:"MAX_CONCURRENT_RECONCILES"`
	KloudliteNamespace                     string `env:"KLOUDLITE_NAMESPACE" default:"kloudlite"`
	ServiceInterceptWebhookServiceSelector map[string]string
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	s, ok := os.LookupEnv("SERVICE_INTERCEPT_WEBHOOK_SERVICE_SELECTOR")
	if ok {
		if err := json.Unmarshal([]byte(s), &ev.ServiceInterceptWebhookServiceSelector); err != nil {
			panic("invalid env-var 'SERVICE_INTERCEPT_WEBHOOK_SERVICE_SELECTOR', must be deserializable into a map[string]string")
		}
	}
	if !ok {
		ev.ServiceInterceptWebhookServiceSelector = map[string]string{"app": "kl-agent-operator"}
		// panic("env-var 'SERVICE_INTERCEPT_WEBHOOK_SERVICE_SELECTOR' not provided")
	}

	return &ev
}
