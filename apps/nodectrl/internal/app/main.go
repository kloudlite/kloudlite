package app

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"kloudlite.io/apps/nodectrl/internal/domain"
	"kloudlite.io/apps/nodectrl/internal/domain/common"
	"kloudlite.io/apps/nodectrl/internal/domain/utils"
	"kloudlite.io/apps/nodectrl/internal/env"
)

var Module = fx.Module("app",
	domain.Module,
	fx.Invoke(
		func(env *env.Env, pc common.ProviderClient, shutdowner fx.Shutdowner, lifecycle fx.Lifecycle) {
			lifecycle.Append(fx.Hook{
				OnStart: func(context.Context) error {

					go func() error {
						ctx := context.Background()
						if err := utils.SetupGetWorkDir(); err != nil {
							return err
						}

						err := func() error {
							switch env.Action {
							case "create":

								fmt.Println("needs to create node")
								if err := pc.NewNode(ctx); err != nil {
									return err
								}
							case "delete":
								fmt.Println("needs to delete node")
								if err := pc.DeleteNode(ctx); err != nil {
									return err
								}

							case "":
								return fmt.Errorf("ACTION not provided, supported actions {create, delete} ")
							default:
								return fmt.Errorf("not supported actions '%s' please provide one of supported action like { create, delete }", env.Action)

							}
							fmt.Println(utils.ColorText("\n🙃 Successfully Exited 🙃\n", 5))
							shutdowner.Shutdown()
							return nil
						}()

						if err != nil {
							fmt.Println(utils.ColorText(fmt.Sprint("\n", "Error: ", err, "\n"), 1))
							return err
						}
						return nil
					}()

					return nil
				},
				OnStop: func(context.Context) error {
					return nil
				},
			})

		},
	),
)
