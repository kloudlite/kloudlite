package graph

import "kloudlite.io/apps/auth/internal/domain"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	d domain.Domain
}

func NewResolver(d domain.Domain) *Resolver {
	return &Resolver{
		d: d,
	}
}
