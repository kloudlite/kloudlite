package domain

import "kloudlite.io/pkg/stripe"

type Stripe interface {
	GetSetupIntent() (string, error)
	CreateCustomer(accountId string, paymentMethodId string) (*stripe.CustomerId, error)
	MakePayment(customerId stripe.CustomerId, paymentMethodId string, amount float64) (*stripe.Payment, error)
}
