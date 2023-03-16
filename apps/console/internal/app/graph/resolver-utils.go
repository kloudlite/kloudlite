package graph

import (
	"context"
	"fmt"

	"kloudlite.io/apps/console/internal/domain"
)

func toConsoleContext(ctx context.Context) domain.ConsoleContext {
	if cc, ok := ctx.Value("kloudlite-ctx").(domain.ConsoleContext); ok {
		return cc
	}
	panic(fmt.Errorf("context values '%s' is missing", "kloudlite-ctx"))
}
