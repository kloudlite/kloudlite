package client

import (
	"github.com/kloudlite/operator/operators/wireguard/apps/multi-cluster/apps/client/env"
	"github.com/kloudlite/operator/operators/wireguard/apps/multi-cluster/mpkg/wg"
	"github.com/kloudlite/operator/pkg/logging"
)

func Run() error {
	env := env.GetEnvOrDie()

	pub, priv, err := wg.GenerateWgKeys()
	if err != nil {
		return err
	}

	l, err := logging.New(&logging.Options{})
	if err != nil {
		return err
	}

	wgc, err := wg.NewClient(env.Interface)
	if err != nil {
		return err
	}

	c := &client{
		logger:     l.WithName("agent"),
		client:     wgc,
		env:        env,
		privateKey: priv,
		publicKey:  pub,
	}

	if err := c.client.Stop(); err != nil {
		c.logger.Error(err)
	}

	defer c.client.Stop()

	return c.start()
}
