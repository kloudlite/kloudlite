package graph

import (
	"github.com/kloudlite/api/apps/auth/internal/domain"
	"github.com/kloudlite/api/apps/auth/internal/env"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	d  domain.Domain
	ev *env.Env
}

func NewResolver(d domain.Domain, ev *env.Env) *Resolver {
	return &Resolver{
		d: d,
		ev: ev,
	}
}
