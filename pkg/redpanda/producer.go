package redpanda

import (
	"context"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/fx"
	"strings"
)

type Producer interface {
	Close()
	Produce(ctx context.Context, topic string, key string, value []byte) error
}

type ProducerImpl struct {
	client *kgo.Client
}

func (p *ProducerImpl) Close() {
	p.client.Close()
}

func (p *ProducerImpl) Produce(ctx context.Context, topic string, key string, value []byte) error {
	record := kgo.KeySliceRecord([]byte(key), value)
	record.Topic = topic
	return p.client.ProduceSync(ctx, record).FirstErr()
}

func NewProducer(brokerHosts string) (Producer, error) {
	client, err := kgo.NewClient(kgo.SeedBrokers(strings.Split(brokerHosts, ",")...))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(context.TODO()); err != nil {
		return nil, err
	}
	return &ProducerImpl{client: client}, nil
}

func NewProducerFx() fx.Option {
	return fx.Module(
		"redis",
		fx.Provide(
			func(client Client) (Producer, error) {
				return NewProducer(client.GetBrokerHosts())
			},
		),
		fx.Invoke(
			func(lf fx.Lifecycle, r Producer) {
				lf.Append(
					fx.Hook{
						OnStop: func(ctx context.Context) error {
							r.Close()
							return nil
						},
					},
				)
			},
		),
	)
}
