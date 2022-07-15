package domain

import (
	"context"
	"kloudlite.io/pkg/stripe"
)

func (d *domainI) MakePayment(customerId stripe.CustomerId, paymentMethodId string, amount float64) (*stripe.Payment, error) {
	return d.stripeCli.NewPayment(customerId, paymentMethodId, amount)
}

func (d *domainI) CreateCustomer(accountId string, paymentMethodId string) (*stripe.CustomerId, error) {
	return d.stripeCli.NewCustomer(accountId, paymentMethodId)
}

func (d *domainI) GetSetupIntent(_ context.Context) (string, error) {
	return d.stripeCli.NewSetupIntent()
}
