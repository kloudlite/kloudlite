package app

import (
	"kloudlite.io/apps/finance/internal/domain"
	"kloudlite.io/pkg/stripe"
)

type pStripe struct {
	stripeCli *stripe.Client
}

func (p pStripe) MakePayment(customerId stripe.CustomerId, paymentMethodId string, amount float64) (*stripe.Payment,
	error) {
	return p.stripeCli.NewPayment(customerId, paymentMethodId, amount)
}

func (p pStripe) CreateCustomer(accountId string, paymentMethodId string) (*stripe.CustomerId, error) {
	return p.stripeCli.NewCustomer(accountId, paymentMethodId)
}

func (p pStripe) GetSetupIntent() (string, error) {
	return p.stripeCli.NewSetupIntent()
}

func NewStripeClient(env *Env) domain.Stripe {
	return &pStripe{stripeCli: stripe.NewClient(env.StripeSecretKey)}
}
