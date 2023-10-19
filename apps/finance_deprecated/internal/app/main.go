package app

import (
	"context"

	"kloudlite.io/constants"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/finance_deprecated/internal/app/graph"
	"kloudlite.io/apps/finance_deprecated/internal/app/graph/generated"
	"kloudlite.io/apps/finance_deprecated/internal/domain"
	"kloudlite.io/apps/finance_deprecated/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/pkg/cache"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

// type Env struct {
// 	CookieDomain    string `env:"COOKIE_DOMAIN" required:"true"`
// 	StripePublicKey string `env:"STRIPE_PUBLIC_KEY" required:"true"`
// 	StripeSecretKey string `env:"STRIPE_SECRET_KEY" required:"true"`
// }

// func (e *WorkloadFinanceConsumerEnv) GetSubscriptionTopics() []string {
// 	return []string{
// 		e.Topic,
// 	}
// }
//
// func (*WorkloadFinanceConsumerEnv) GetConsumerGroupId() string {
// 	return "console-workload-finance-consumer-2"
// }

type AuthCacheClient cache.Client

var Module = fx.Module(
	"application",
	// config.EnvFx[Env](),
	repos.NewFxMongoRepo[*domain.Account]("accounts", "acc", domain.AccountIndexes),
	repos.NewFxMongoRepo[*domain.AccountBilling]("account_billings", "accbill", domain.BillableIndexes),
	repos.NewFxMongoRepo[*domain.BillingInvoice]("account_invoices", "inv", domain.BillingInvoiceIndexes),
	cache.NewFxRepo[*domain.AccountInviteToken](),
	IAMClientFx,
	ConsoleClientFx,
	ContainerRegistryFx,
	AuthClientFx,
	CommsClientFx,
	fx.Invoke(
		func(server *fiber.App, d domain.Domain, env *env.Env, cacheClient AuthCacheClient) {
			gqlConfig := generated.Config{Resolvers: graph.NewResolver(d)}

			gqlConfig.Directives.IsLoggedIn = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				cc := domain.FinanceContext{Context: ctx, UserId: sess.UserId}
				return next(context.WithValue(ctx, "kl-finance-ctx", cc))
			}

			schema := generated.NewExecutableSchema(gqlConfig)
			httpServer.SetupGQLServer(
				server,
				schema,
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					constants.CookieName,
					env.CookieDomain,
					constants.CacheSessionPrefix,
				),
			)
		},
	),

	// config.EnvFx[WorkloadFinanceConsumerEnv](),
	// fx.Provide(
	// 	func(env *Env) *stripe.Client {
	// 		return stripe.NewClient(env.StripeSecretKey)
	// 	},
	// ),
	fx.Invoke(
		func(server *fiber.App) {
			// server.Get(
			//	"/stripe/get-setup-intent", func(ctx *fiber.Ctx) error {
			//		intentClientSecret, err := ds.GetSetupIntent()
			//		if err != nil {
			//			return err
			//		}
			//		return ctx.JSON(
			//			map[string]any{
			//				"client-secret": intentClientSecret,
			//			},
			//		)
			//	},
			// )

			// server.Post(
			//	"/stripe/create-customer", func(ctx *fiber.Ctx) error {
			//		var j struct {
			//			AccountId       string `json:"accountId"`
			//			PaymentMethodId string `json:"paymentMethodId"`
			//		}
			//		if err := json.Unmarshal(ctx.Body(), &j); err != nil {
			//			return err
			//		}
			//		customer, err := ds.CreateCustomer(j.AccountId, j.PaymentMethodId)
			//		if err != nil {
			//			return errors.NewEf(err, "creating customer")
			//		}
			//		//payment, err := ds.MakePayment(*customer, j.PaymentMethodId, 20000)
			//		//if err != nil {
			//		//	return errors.NewEf(err, "making initial payment")
			//		//}
			//		return ctx.JSON(
			//			map[string]any{
			//				"customerId":   *customer,
			//				"init-payment": payment,
			//			},
			//		)
			//	},
			// )
		},
	),

	fx.Provide(fxFinanceGrpcServer),
	fx.Invoke(
		func(server *grpc.Server, financeServer finance.FinanceServer) {
			finance.RegisterFinanceServer(server, financeServer)
		},
	),
	domain.Module,
)
