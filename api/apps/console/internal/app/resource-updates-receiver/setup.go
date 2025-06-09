package resource_updates_receiver

import (
	"context"
	"log/slog"
	"sync"

	"github.com/kloudlite/api/apps/console/internal/domain"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"
)

type StartArgs struct {
	ResourceUpdateConsumer
	ErrorOnApplyConsumer
	domain.Domain
	*slog.Logger
}

func Start(args StartArgs) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		ProcessResourceUpdates(args.ResourceUpdateConsumer, args.Domain, args.Logger)
	}()

	go func() {
		defer wg.Done()
		ProcessErrorOnApply(args.ErrorOnApplyConsumer, args.Domain, args.Logger)
	}()

	wg.Wait()
}

type StopArgs struct {
	ResourceUpdateConsumer
	ErrorOnApplyConsumer
}

func Stop(ctx context.Context, args StopArgs) error {
	errg := new(errgroup.Group)

	errg.Go(func() error {
		return args.ResourceUpdateConsumer.Stop(ctx)
	})

	errg.Go(func() error {
		return args.ErrorOnApplyConsumer.Stop(ctx)
	})

	return errg.Wait()
}

var Module = fx.Module(
	"resource_updates_receiver",
	fx.Invoke(
		func(lf fx.Lifecycle, ruc ResourceUpdateConsumer, eoa ErrorOnApplyConsumer, d domain.Domain, logger *slog.Logger) {
			lf.Append(fx.Hook{
				OnStart: func(context.Context) error {
					go Start(StartArgs{
						ResourceUpdateConsumer: ruc,
						ErrorOnApplyConsumer:   eoa,
						Domain:                 d,
						Logger:                 logger,
					})
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return Stop(ctx, StopArgs{
						ResourceUpdateConsumer: ruc,
						ErrorOnApplyConsumer:   eoa,
					})
				},
			})
		}),
)
