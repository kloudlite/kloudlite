package graph

import (
	"github.com/kloudlite/api/apps/container-registry/internal/domain"
	"github.com/kloudlite/api/apps/container-registry/internal/env"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Domain domain.Domain
	Env    *env.Env
}
