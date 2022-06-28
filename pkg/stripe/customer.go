package stripe

import (
	"fmt"
	"github.com/stripe/stripe-go"
)

type Customer struct {
	StripeCustomerId string
}

func (c *Client) NewCustomer(accountId string, paymentMethodId string) (*Customer, error) {
	customer, err := c.stripe.Customers.New(
		&stripe.CustomerParams{
			Name:          &accountId,
			Description:   stripe.String(fmt.Sprintf("kloudlite accountId=%s", accountId)),
			PaymentMethod: &paymentMethodId,
		},
	)
	if err != nil {
		return nil, err
	}
	return &Customer{StripeCustomerId: customer.ID}, nil
}

func (c *Client) DeleteCustomer(cus *Customer) error {
	if _, err := c.stripe.Customers.Del(cus.StripeCustomerId, &stripe.CustomerParams{}); err != nil {
		return err
	}
	return nil
}
