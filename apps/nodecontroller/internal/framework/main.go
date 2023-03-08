package framework

import (
	"fmt"
	"os"

	"go.uber.org/fx"
	"kloudlite.io/apps/nodecontroller/internal/app"
	"kloudlite.io/apps/nodecontroller/internal/domain"
)

var Module = fx.Module(
	"framework",
	app.Module,
	fx.Invoke(func(d domain.Domain, shutdowner fx.Shutdowner) {
		err := d.StartJob()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			shutdowner.Shutdown()
		}
	}),
)
