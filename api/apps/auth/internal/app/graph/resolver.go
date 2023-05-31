package graph

import (
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/apps/auth/internal/env"
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
