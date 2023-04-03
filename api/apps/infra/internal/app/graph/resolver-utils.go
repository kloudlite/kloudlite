package graph

import (
	"context"
	"fmt"
	"kloudlite.io/apps/infra/internal/domain"
)

func toInfraContext(ctx context.Context) domain.InfraContext {
	if d, ok := ctx.Value("infra-ctx").(domain.InfraContext); ok {
		return d
	}
	panic(fmt.Errorf("infra context not found in gql context"))
}
