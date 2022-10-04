package redpanda

import (
	"context"
	"github.com/twmb/franz-go/pkg/kgo"
	"strings"
)

type Producer interface {
	Ping(ctx context.Context) error
	Close()
	Produce(ctx context.Context, topic, key string, value []byte) (*ProducerOutput, error)
}

type producer struct {
	client *kgo.Client
}

func (p *producer) Ping(ctx context.Context) error {
	return p.client.Ping(ctx)
}

func (p *producer) Close() {
	p.client.Close()
}

// func (p *producer) Produce(ctx context.Context, topic string, key string, value []byte) error {
// 	record := kgo.KeySliceRecord([]byte(key), value)
// 	record.Topic = topic
// 	return p.client.ProduceSync(ctx, record).FirstErr()
// }

func (p *producer) Produce(ctx context.Context, topic string, key string, value []byte) (*ProducerOutput, error) {
	record := kgo.KeySliceRecord(
		func() []byte {
			if key == "" {
				return nil
			}
			return []byte(key)
		}(), value,
	)
	record.Topic = topic
	sync, err := p.client.ProduceSync(ctx, record).First()
	if err != nil {
		return nil, err
	}
	return &ProducerOutput{
		Key:        sync.Key,
		Timestamp:  sync.Timestamp,
		Topic:      sync.Topic,
		Partition:  sync.Partition,
		ProducerId: sync.ProducerID,
		Offset:     sync.Offset,
	}, nil
}

func NewProducer(brokerHosts string, producerOpts ProducerOpts) (Producer, error) {
	opts := make([]kgo.Opt, 0, 2)
	opts = append(opts, kgo.SeedBrokers(strings.Split(brokerHosts, ",")...))

	saslOpt, err := parseSASLAuth(producerOpts.SASLAuth)
	if err != nil {
		return nil, err
	}
	opts = append(opts, saslOpt)

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &producer{client: client}, nil
}
