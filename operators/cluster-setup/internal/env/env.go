package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`
	// OperatorTemplatesDir    string        `env:"OPERATOR_TEMPLATES_DIR" required:"true"`

	ReleasesApiEndpoint string `env:"RELEASES_API_ENDPOINT" required:"true"`
	ReleasesApiUsername string `env:"RELEASES_API_USERNAME" required:"true"`
	ReleasesApiPassword string `env:"RELEASES_API_PASSWORD" required:"true"`
	ReleaseVersion      string `env:"RELEASE_VERSION" required:"true"`

	HelmChartsDir string `env:"HELM_CHARTS_DIR" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
