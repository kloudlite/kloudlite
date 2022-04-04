package config

import (
	"github.com/codingconcepts/env"
	"github.com/joho/godotenv"
)

func LoadConfigFromEnv(b interface{}) error {
	return env.Set(b)
}

func LoadDotEnv() error {
	return godotenv.Load()
}
