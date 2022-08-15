package app

import (
	"context"
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
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
	"kloudlite.io/pkg/stripe"
	"strconv"
	"time"
)

type Env struct {
	CookieDomain    string `env:"COOKIE_DOMAIN" required:"true"`
	StripePublicKey string `env:"STRIPE_PUBLIC_KEY" required:"true"`
	StripeSecretKey string `env:"STRIPE_SECRET_KEY" required:"true"`
}

type WorkloadFinanceConsumerEnv struct {
	Topic string `env:"KAFKA_WORKLOAD_FINANCE_TOPIC"`
}

func (e *WorkloadFinanceConsumerEnv) GetSubscriptionTopics() []string {
	return []string{
		e.Topic,
	}
}

func (*WorkloadFinanceConsumerEnv) GetConsumerGroupId() string {
	return "console-workload-finance-consumer-2"
}

type AuthCacheClient cache.Client

var Module = fx.Module(
	"application",
	config.EnvFx[Env](),
	repos.NewFxMongoRepo[*domain.Account]("accounts", "acc", domain.AccountIndexes),
	repos.NewFxMongoRepo[*domain.AccountBilling]("account_billings", "accbill", domain.BillableIndexes),
	repos.NewFxMongoRepo[*domain.BillingInvoice]("account_invoices", "inv", domain.BillingInvoiceIndexes),
	cache.NewFxRepo[*domain.AccountInviteToken](),
	IAMClientFx,
	ConsoleClientFx,
	AuthClientFx,
	CommsClientFx,
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

	config.EnvFx[WorkloadFinanceConsumerEnv](),
	redpanda.NewConsumerFx[*WorkloadFinanceConsumerEnv](),
	fx.Invoke(func(d domain.Domain, consumer redpanda.Consumer) {
		consumer.StartConsuming(func(msg []byte, timeStamp time.Time, offset int64) error {
			var e domain.BillingEvent
			err := json.Unmarshal(msg, &e)
			if err != nil {
				fmt.Println(err)
				return err
			}
			err = d.TriggerBillingEvent(
				context.TODO(),
				repos.ID(e.Metadata.AccountId),
				repos.ID(e.Metadata.ResourceId),
				repos.ID(e.Metadata.ProjectId),
				(func() string {
					fmt.Println(e.Stage)
					if e.Stage == "EXISTS" {
						return "exists"
					} else {
						return "end"
					}
				})(),
				func() []domain.Billable {
					billables := make([]domain.Billable, 0)
					for _, i := range e.Billing.Items {
						if q, err := strconv.ParseFloat(i.PlanQ, 32); err == nil {
							billables = append(billables, domain.Billable{
								ResourceType: i.Type,
								Plan:         i.Plan,
								Quantity:     q,
								Count:        i.Count,
								IsShared:     i.IsShared == "true",
							})
						}
					}
					return billables
				}(),
				timeStamp,
			)
			fmt.Println(err)
			return err
		})
	}),

	fx.Provide(func(env *Env) *stripe.Client {
		return stripe.NewClient(env.StripeSecretKey)
	}),
	fx.Invoke(
		func(server *fiber.App) {
			//server.Get(
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
			//)

			//server.Post(
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
			//)
		},
	),

	domain.Module,
)
