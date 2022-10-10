package redpanda

import "go.uber.org/fx"

type Client interface {
	GetBrokerHosts() string
	GetKafkaSASLAuth() *KafkaSASLAuth
}

type ClientImpl struct {
	brokerHosts string
	saslAuth    *KafkaSASLAuth
}

func (c *ClientImpl) GetKafkaSASLAuth() *KafkaSASLAuth {
	return c.saslAuth
}

func (c *ClientImpl) GetBrokerHosts() string {
	return c.brokerHosts
}

func NewClient(brokers string, auth *KafkaSASLAuth) Client {
	return &ClientImpl{
		brokerHosts: brokers,
		saslAuth:    auth,
	}
}

type ClientConfig interface {
	GetBrokers() (brokers string)
	GetKafkaSASLAuth() *KafkaSASLAuth
}

func NewClientFx[T ClientConfig]() fx.Option {
	return fx.Module(
		"redpanda",
		fx.Provide(func(env T) Client {
			return NewClient(env.GetBrokers(), env.GetKafkaSASLAuth())
		}),
	)
}
