package graph

import (
	"context"
	"fmt"
	"kloudlite.io/apps/container-registry/internal/domain"
)

func toRegistryContext(ctx context.Context) domain.RegistryContext {
	if cc, ok := ctx.Value("kloudlite-ctx").(domain.RegistryContext); ok {
		return cc
	}
	panic(fmt.Errorf("context values '%s' is missing", "kloudlite-ctx"))
}
