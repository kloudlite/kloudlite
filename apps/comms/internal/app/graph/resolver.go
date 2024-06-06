package graph

import (
	"github.com/kloudlite/api/apps/comms/internal/domain"
	"github.com/kloudlite/api/apps/comms/internal/env"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Domain domain.Domain
	Env    *env.Env
}
