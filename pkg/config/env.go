package config

import (
	"fmt"

	"github.com/codingconcepts/env"
)

func LoadEnv[T any]() func() (T, error) {
	return func() (T, error) {
		var envC T
		err := env.Set(&envC)
		if err != nil {
			var e T
			return e, fmt.Errorf("not able to load ENV: %v", err)
		}
		return envC, err
	}
}
