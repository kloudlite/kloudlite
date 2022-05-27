package config

import (
	"fmt"
	"go.uber.org/fx"

	"github.com/codingconcepts/env"
)

func LoadEnv[T any]() func() (*T, error) {
	return func() (*T, error) {
		var x T
		err := env.Set(&x)
		if err != nil {
			return nil, fmt.Errorf("not able to load ENV: %v", err)
		}
		return &x, err
	}
}

func EnvFx[T any]() fx.Option {
	return fx.Module(
		"env",
		fx.Provide(LoadEnv[T]()),
	)
}
