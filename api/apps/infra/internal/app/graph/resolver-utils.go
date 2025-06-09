package graph

import (
	"context"
	"github.com/kloudlite/api/apps/infra/internal/domain"
	"github.com/kloudlite/api/pkg/errors"
)

func toInfraContext(ctx context.Context) (domain.InfraContext, error) {
	if d, ok := ctx.Value("infra-ctx").(domain.InfraContext); ok {
		return d, nil
	}
	return domain.InfraContext{}, errors.Newf("infra context not found in gql context")
}
