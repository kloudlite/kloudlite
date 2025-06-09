package nats

import (
	"context"
	"log/slog"
	"time"

	"github.com/kloudlite/api/pkg/errors"

	"github.com/nats-io/nats.go"
)

type Client struct {
	Conn   *nats.Conn
	logger *slog.Logger
}

// Close implements Client.
func (nc *Client) Close(ctx context.Context) error {
	nc.Conn.Close()
	return nil
}

// expose these nats types, so that we can use them directly, without having to import nats.io/nats.go
type (
	Options nats.Options
)

type ClientOpts struct {
	CrdentialsFile string
	// Options

	Name string
	// https://pkg.go.dev/github.com/nats-io/nats.go#Options
	Servers []string
	Logger  *slog.Logger

	DisconnectedCB func()
	ReconnectedCB  func()
	ConnectedCB    func()
	ClosedCB       func()
}

func NewClient(url string, opts ClientOpts) (*Client, error) {
	if opts.Name == "" {
		return nil, errors.Newf("opts.name is required")
	}

	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	connectOpts := []nats.Option{
		func(o *nats.Options) error {
			// *o = nats.Options(opts.Options)
			*o = nats.Options{
				Url: url,
				Servers: func() []string {
					// INFO: without setting this servers with the url, i am not able to connect to nats either hosted at synadia cloud, or my own helm installed cluster
					serverUrlExists := false
					for i := range opts.Servers {
						if url == opts.Servers[i] {
							serverUrlExists = true
						}
					}

					if !serverUrlExists {
						opts.Servers = append(opts.Servers, url)
					}

					return opts.Servers
				}(),
				Name:           opts.Name,
				Verbose:        false,
				Pedantic:       false,
				Secure:         false,
				AllowReconnect: true,
				MaxReconnect:   -1,
				ReconnectWait:  3 * time.Second,
				PingInterval:   3 * time.Second,
				MaxPingsOut:    0,
				ClosedCB: func(*nats.Conn) {
					if opts.ClosedCB != nil {
						opts.ClosedCB()
						return
					}
					opts.Logger.Warn("connection closed with nats server", "name", opts.Name)
				},
				DisconnectedCB: func(*nats.Conn) {
					if opts.DisconnectedCB != nil {
						opts.DisconnectedCB()
						return
					}
					opts.Logger.Warn("disconnected with nats server", "name", opts.Name)
				},
				ConnectedCB: func(*nats.Conn) {
					if opts.ConnectedCB != nil {
						opts.ConnectedCB()
						return
					}
					opts.Logger.Info("connected to nats server", "name", opts.Name)
				},
				ReconnectedCB: func(*nats.Conn) {
					if opts.ReconnectedCB != nil {
						opts.ReconnectedCB()
						return
					}
					opts.Logger.Info("re-connected to nats server", "name", opts.Name)
				},
				DiscoveredServersCB: func(c *nats.Conn) {
					opts.Logger.Info("discovered additional nats servers", "servers", c.DiscoveredServers())
				},
				AsyncErrorCB: func(_ *nats.Conn, sub *nats.Subscription, err error) {
					opts.Logger.Warn("got async error for", "name", opts.Name, "subject", sub.Subject, "err", err)
				},
				RetryOnFailedConnect: true,
				Compression:          true,
			}

			return nil
		},
	}

	if opts.CrdentialsFile != "" {
		connectOpts = append(connectOpts, nats.UserCredentials(opts.CrdentialsFile))
	}

	nc, err := nats.Connect(url, connectOpts...)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &Client{
		Conn:   nc,
		logger: opts.Logger,
	}, nil
}
