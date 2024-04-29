package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	ConfigPath string `env:"CONFIG_PATH" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
