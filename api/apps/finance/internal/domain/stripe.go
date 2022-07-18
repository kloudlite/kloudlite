package domain

import (
	"context"
	"fmt"
	"kloudlite.io/pkg/repos"
	"kloudlite.io/pkg/stripe"
)

func (d *domainI) Test(ctx context.Context, accountId repos.ID) error {
	acc, err := d.accountRepo.FindById(ctx, accountId)
	payment, err := d.stripeCli.NewPayment(stripe.CustomerId(acc.Billing.StripeCustomerId), acc.Billing.PaymentMethodId, 50)
	fmt.Println(payment)
	return err
}

//func (d *domainI) MakePayment(customerId stripe.CustomerId, paymentMethodId string, amount float64) (*stripe.Payment, error) {
//	return d.stripeCli.NewPayment(customerId, paymentMethodId, amount)
//}

func (d *domainI) GetSetupIntent(_ context.Context) (string, error) {
	return d.stripeCli.NewSetupIntent()
}
