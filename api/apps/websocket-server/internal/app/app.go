package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain"
	"github.com/kloudlite/api/apps/websocket-server/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
)

type AuthCacheClient kv.Client

type AuthClient grpc.Client

type (
	ContainerRegistryClient grpc.Client
	IAMClient               grpc.Client
)

var Module = fx.Module("app",

	// grpc clients
	fx.Provide(func(conn IAMClient) iam.IAMClient {
		return iam.NewIAMClient(conn)
	}),

	domain.Module,

	fx.Invoke(
		func(server httpServer.Server, d domain.Domain, env *env.Env,
			logr logging.Logger,
			sessionRepo kv.Repo[*common.AuthSession],
		) {
			a := server.Raw()

			a.Use(httpServer.NewReadSessionMiddleware(sessionRepo, constants.CookieName, constants.CacheSessionPrefix))

			// Web socket route
			a.Use("/ws", func(c *fiber.Ctx) error {
				if websocket.IsWebSocketUpgrade(c) {
					return c.Next()
				}
				return fiber.ErrUpgradeRequired
			})

			a.Use("/ws", func(c *fiber.Ctx) error {
				ctx := c.Context()

				return websocket.New(func(sockConn *websocket.Conn) {
					if err := d.HandleWebSocket(ctx, sockConn); err != nil {
						logr.Errorf(err, "while handling websocket for resource update")
					}
				})(c)
			})

			a.Get("/healthy", func(c *fiber.Ctx) error {
				return c.SendString("OK")
			})

			a.All("*", func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusNotFound)
			})
		},
	),
)
