package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	// example: "http://localhost:8001"
	ServerAddr string `env:"SERVER_ADDR" required:"true"`
	KubeDns    string `env:"KUBE_DNS_IP" required:"true"`
	MyIp       string `env:"MY_IP_ADDRESS" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
