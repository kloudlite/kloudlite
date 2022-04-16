package config

import (
	"fmt"

	"github.com/codingconcepts/env"
)

type Samole struct {
}

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
