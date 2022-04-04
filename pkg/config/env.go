package config

import (
	"flag"
	"github.com/codingconcepts/env"
)

type BaseEnv struct {
	IsDev bool `env:"IS_DEV", required:"true"`
}

func (b *BaseEnv) Load() {
	if err := env.Set(b); err != nil {
		panic(err)
	}
	isDev := flag.Bool("dev", false, "isDevelopment")
	flag.Parse()
	b.IsDev = *isDev
}
