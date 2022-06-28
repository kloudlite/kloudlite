package stripe

import (
	"github.com/stripe/stripe-go/client"
)

type Client struct {
	stripe *client.API
}

func NewClient(pk string) *Client {
	cli := client.New(pk, nil)
	return &Client{stripe: cli}
}
