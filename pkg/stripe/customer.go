package stripe

import (
	"fmt"
	"github.com/stripe/stripe-go"
	fn "kloudlite.io/pkg/functions"
)

type CustomerId string

func (c *CustomerId) Str() string {
	return string(*c)
}

func (c *CustomerId) StrP() *string {
	return fn.New(string(*c))
}

func (c *Client) NewCustomer(accountId string, paymentMethodId string) (*CustomerId, error) {
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
	return fn.New(CustomerId(customer.ID)), nil
}

func (c *Client) DeleteCustomer(cus CustomerId) error {
	if _, err := c.stripe.Customers.Del(cus.Str(), &stripe.CustomerParams{}); err != nil {
		return err
	}
	return nil
}
