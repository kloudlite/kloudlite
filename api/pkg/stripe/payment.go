package stripe

import (
	"github.com/stripe/stripe-go"
	"math"
)

// Note: [Stripe Docs Ref](https://stripe.com/docs/api/charges)

type CardId string

type Payment struct {
	IntentId     string
	ClientSecret string
	MethodId     string
	Status       bool
}

func (c *Client) NewPayment(cus CustomerId, paymentMethodId string, amount float64) (*Payment, error) {
	params := stripe.PaymentIntentParams{
		Amount:        stripe.Int64(int64(math.Round(amount))),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Customer:      cus.StrP(),
		PaymentMethod: stripe.String(paymentMethodId),
		Confirm:       stripe.Bool(true),
		OffSession:    stripe.Bool(true),
	}

	intent, err := c.stripe.PaymentIntents.New(&params)
	if err != nil {
		if sErr, ok := err.(*stripe.Error); ok {
			if sErr.Code == stripe.ErrorCodeAuthenticationRequired {
				return &Payment{
					IntentId:     sErr.PaymentIntent.ID,
					ClientSecret: sErr.PaymentIntent.ClientSecret,
					MethodId:     sErr.PaymentMethod.ID,
				}, err
			}
		}
		return nil, err
	}

	return &Payment{
		IntentId: intent.ID,
		MethodId: intent.PaymentMethod.ID,
		Status:   intent.Status == stripe.PaymentIntentStatusSucceeded,
	}, nil
}
