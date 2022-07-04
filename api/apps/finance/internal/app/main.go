package app

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"kloudlite.io/apps/finance/internal/app/graph"
	"kloudlite.io/apps/finance/internal/app/graph/generated"
	"kloudlite.io/apps/finance/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/errors"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	CookieDomain    string `env:"COOKIE_DOMAIN" required:"true"`
	StripePublicKey string `env:"STRIPE_PUBLIC_KEY" required:"true"`
	StripeSecretKey string `env:"STRIPE_SECRET_KEY" required:"true"`
}

type AuthCacheClient cache.Client

var Module = fx.Module(
	"application",
	config.EnvFx[Env](),
	repos.NewFxMongoRepo[*domain.Account]("accounts", "acc", domain.AccountIndexes),
	repos.NewFxMongoRepo[*domain.Billable]("billables", "bill", domain.BillableIndexes),
	cache.NewFxRepo[*domain.AccountInviteToken](),
	CiClientFx,
	IAMClientFx,
	ConsoleClientFx,
	AuthClientFx,
	fx.Invoke(
		func(server *fiber.App, d domain.Domain, env *Env, cacheClient AuthCacheClient) {
			schema := generated.NewExecutableSchema(
				generated.Config{Resolvers: graph.NewResolver(d)},
			)
			httpServer.SetupGQLServer(
				server,
				schema,
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					common.CookieName,
					env.CookieDomain,
					common.CacheSessionPrefix,
				),
			)
		},
	),

	fx.Provide(NewStripeClient),
	fx.Invoke(
		func(server *fiber.App, ds domain.Stripe) {
			server.Get(
				"/stripe/get-setup-intent", func(ctx *fiber.Ctx) error {
					intentClientSecret, err := ds.GetSetupIntent()
					if err != nil {
						return err
					}
					return ctx.JSON(
						map[string]any{
							"client-secret": intentClientSecret,
						},
					)
				},
			)

			server.Post(
				"/stripe/create-customer", func(ctx *fiber.Ctx) error {
					var j struct {
						AccountId       string `json:"accountId"`
						PaymentMethodId string `json:"paymentMethodId"`
					}
					if err := json.Unmarshal(ctx.Body(), &j); err != nil {
						return err
					}
					customer, err := ds.CreateCustomer(j.AccountId, j.PaymentMethodId)
					if err != nil {
						return errors.NewEf(err, "creating customer")
					}

					payment, err := ds.MakePayment(*customer, j.PaymentMethodId, 20000)
					if err != nil {
						return errors.NewEf(err, "making initial payment")
					}

					fmt.Printf("Payment: %+v\n", payment)

					return ctx.JSON(
						map[string]any{
							"customerId":   *customer,
							"init-payment": payment,
						},
					)
				},
			)
		},
	),

	domain.Module,
)
