package nats

import (
	"context"

	"github.com/nats-io/nats.go"
)

type Client struct {
	Conn *nats.Conn
}

// Close implements Client.
func (nc *Client) Close(ctx context.Context) error {
	nc.Conn.Close()
	return nil
}

type Options nats.Options

type ClientOpts struct {
	CrdeentialsFile string
	Options
}

func NewClient(url string, opts ClientOpts) (*Client, error) {
	connectOpts := []nats.Option{
		func(o *nats.Options) error {
			*o = nats.Options(opts.Options)
			return nil
		},
	}

	if opts.CrdeentialsFile != "" {
		connectOpts = append(connectOpts, nats.UserCredentials(opts.CrdeentialsFile))
	}

	nc, err := nats.Connect(url, connectOpts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		Conn: nc,
	}, nil
}
