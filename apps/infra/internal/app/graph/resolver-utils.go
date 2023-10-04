package graph

import (
	"context"
	"fmt"
	"kloudlite.io/apps/infra/internal/domain"
)

func toInfraContext(ctx context.Context) (domain.InfraContext, error) {
	if d, ok := ctx.Value("infra-ctx").(domain.InfraContext); ok {
		return d, nil
	}
	return domain.InfraContext{}, fmt.Errorf("infra context not found in gql context")
}
