package config

import (
	"fmt"
	"github.com/codingconcepts/env"
	"github.com/joho/godotenv"
)

func LoadConfigFromEnv(b interface{}) error {
	return env.Set(b)
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file: %v", err))
	}
}
