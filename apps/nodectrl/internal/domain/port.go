package domain

import (
	"kloudlite.io/apps/nodectrl/internal/env"
)

type Domain interface {
	GetEnv() *env.Env
}
