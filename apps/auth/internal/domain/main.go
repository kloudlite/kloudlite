package domain

import (
	"context"

	"go.uber.org/fx"
	"kloudlite.io/pkg/repos"
)

var Module = fx.Module(
	"domain",
	fx.Provide(fxDomain),
	fx.Invoke(func(lf fx.Lifecycle, repo repos.DbRepo[*User]) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				repo.FindOne(ctx, repos.Filter{
					"email": "nxtcoder17@gmail.com",
				})
				return nil
			},
		})
	}),
)
