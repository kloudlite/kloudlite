package messaging

import "go.uber.org/fx"

type KafkaClient interface {
	GetBrokers() string
}

type KafkaClientOptions interface {
	GetBrokers() string
}

type kafkaClient struct {
	Brokers string
}

func (k *kafkaClient) GetBrokers() string {
	return k.Brokers
}

func NewKafkaClient(brokers string) KafkaClient {
	return &kafkaClient{
		Brokers: brokers,
	}
}

func NewKafkaClientFx[T KafkaClientOptions]() fx.Option {
	return fx.Module("kafka", fx.Provide(func(env T) KafkaClient {
		return NewKafkaClient(env.GetBrokers())
	}))
}
