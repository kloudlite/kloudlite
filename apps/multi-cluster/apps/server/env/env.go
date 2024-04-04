package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	// example: ":8001"
	Addr string `env:"ADDR" required:"true"`
	// example: "./examples/server.yml"
	ConfigPath string `env:"CONFIG_PATH" required:"true"`
	Endpoint string `env:"ENDPOINT" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
