package redpanda

import (
	"context"
	"github.com/twmb/franz-go/pkg/kgo"
	"strings"
)

type Producer struct {
	client *kgo.Client
}

func (p *Producer) Close() {
	p.client.Close()
}

func (p *Producer) Produce(ctx context.Context, topic string, key string, value []byte) error {
	record := kgo.KeySliceRecord([]byte(key), value)
	record.Topic = topic
	return p.client.ProduceSync(ctx, record).FirstErr()
}

func NewProducer(brokerHosts string) (*Producer, error) {
	client, err := kgo.NewClient(kgo.SeedBrokers(strings.Split(brokerHosts, ",")...))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(context.TODO()); err != nil {
		return nil, err
	}
	return &Producer{client: client}, nil
}
