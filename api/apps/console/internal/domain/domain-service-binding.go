package domain

import (
	"context"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/pkg/repos"
)

type ServiceBindingDomain interface {
	FindServiceBindingByHostname(ctx context.Context, accountName string, hostname string) (*entities.ServiceBinding, error)
}

type svcBindingDomain struct {
	repo repos.DbRepo[*entities.ServiceBinding]
}

// FindServiceBindingByHostname implements ServiceBindingDomain.
func (s *svcBindingDomain) FindServiceBindingByHostname(ctx context.Context, accountName string, hostname string) (*entities.ServiceBinding, error) {
	return s.repo.FindOne(ctx, repos.Filter{
		fc.AccountName:                accountName,
		fc.ServiceBindingSpecHostname: hostname,
	})
}

func NewSvcBindingDomain(svcBindingRepo repos.DbRepo[*entities.ServiceBinding]) ServiceBindingDomain {
	return &svcBindingDomain{
		repo: svcBindingRepo,
	}
}
