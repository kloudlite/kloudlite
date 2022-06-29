package redpanda

import "go.uber.org/fx"

type Client interface {
	GetBrokerHosts() string
}

type ClientImpl struct {
	brokerHosts string
}

func (c *ClientImpl) GetBrokerHosts() string {
	return c.brokerHosts
}

func NewClient(brokers string) Client {
	return &ClientImpl{
		brokerHosts: brokers,
	}
}

type ClientConfig interface {
	GetBrokers() (brokers string)
}

func NewClientFx[T ClientConfig]() fx.Option {
	return fx.Module(
		"redpanda",
		fx.Provide(func(env T) Client {
			return NewClient(env.GetBrokers())
		}),
	)
}
