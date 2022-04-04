package config

import (
	"github.com/codingconcepts/env"
)

func LoadEnv(b interface{}) error {
	return env.Set(b)
}
