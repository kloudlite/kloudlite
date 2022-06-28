package stripe

import (
	"github.com/stripe/stripe-go"
)

func (c *Client) NewSetupIntent() (clientSecret string, err error) {
	intentParams := stripe.SetupIntentParams{
		PaymentMethodTypes: []*string{
			stripe.String(string(stripe.PaymentMethodTypeCard)),
		},
	}
	intent, err := c.stripe.SetupIntents.New(&intentParams)
	if err != nil {
		return "", err
	}
	return intent.ClientSecret, nil
}
