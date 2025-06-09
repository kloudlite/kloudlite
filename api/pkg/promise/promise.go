package promise

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type AsyncFn func(ctx context.Context) error

func All(parentCtx context.Context, fns ...AsyncFn) error {
	g, ctx := errgroup.WithContext(parentCtx)
	for i := range fns {
		idx := i
		g.Go(func() error {
			return fns[idx](ctx)
		})
	}

	return g.Wait()
}
