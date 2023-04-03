package main

import (
	"context"
	"flag"

	"go.uber.org/fx"
	"kloudlite.io/apps/finance.old/internal/domain"
	"kloudlite.io/apps/finance.old/internal/framework"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
)

func main() {
	accountId := flag.String("dev", "", "--accountId")
	flag.Parse()
	if accountId == nil {
		panic("accountId is required")
	}
	fx.New(
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "console", Dev: false})
			},
		),
		framework.Module,
		fx.Invoke(
			func(d domain.Domain) {
				d.GenerateBillingInvoice(context.Background(), repos.ID(*accountId))
			},
		),
	).Start(context.TODO())
}
