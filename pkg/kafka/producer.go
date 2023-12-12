package kafka

import (
	"context"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/twmb/franz-go/pkg/kgo"
	"strings"
	"time"
)

type MessageArgs struct {
	Key     []byte
	Headers map[string][]byte
}

type Producer interface {
	Ping(ctx context.Context) error
	Close()

	Produce(ctx context.Context, topic string, value []byte, args MessageArgs) (*ProducerOutput, error)

	LifecycleOnStart(ctx context.Context) error
	LifecycleOnStop(ctx context.Context) error
}

type producer struct {
	client *kgo.Client
	logger logging.Logger
}

// LifecycleOnStart implements Producer.
func (p *producer) LifecycleOnStart(ctx context.Context) error {
	p.logger.Debugf("producer is about to ping kafka brokers")
	if err := p.Ping(ctx); err != nil {
		return err
	}
	p.logger.Infof("producer is connected to kafka brokers")
	return nil
}

// LifecycleOnStop implements Producer.
func (p *producer) LifecycleOnStop(context.Context) error {
	p.logger.Debugf("producer is about to be closed")
	p.Close()
	p.logger.Infof("producer is closed")
	return nil
}

func (p *producer) Ping(ctx context.Context) error {
	return p.client.Ping(ctx)
}

func (p *producer) Close() {
	p.client.Close()
}

type ProducerOutput struct {
	Key        []byte    `json:"key,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Topic      string    `json:"topic"`
	Partition  int32     `json:"partition,omitempty"`
	ProducerId int64     `json:"producerId,omitempty"`
	Offset     int64     `json:"offset"`
}

func (p *producer) Produce(ctx context.Context, topic string, value []byte, args MessageArgs) (*ProducerOutput, error) {
	key := args.Key
	if string(key) == "" {
		key = nil
	}

	record := kgo.KeySliceRecord(key, value)
	record.Topic = topic

	headers := make([]kgo.RecordHeader, 0, len(args.Headers))
	for k, v := range args.Headers {
		headers = append(headers, kgo.RecordHeader{Key: k, Value: v})
	}
	record.Headers = headers

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

type ProducerOpts struct {
	Logger logging.Logger
}

func NewProducer(conn Conn, opts ProducerOpts) (Producer, error) {
	if opts.Logger == nil {
		var err error
		opts.Logger, err = logging.New(&logging.Options{Name: "kafka-producer"})
		if err != nil {
			return nil, err
		}
	}

	kgoOpts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(conn.GetBrokers(), ",")...),
	}
	saslOpt, err := parseSASLAuth(conn.GetSASLAuth())
	if err != nil {
		return nil, err
	}

	if saslOpt != nil {
		kgoOpts = append(kgoOpts, saslOpt)
	}

	client, err := kgo.NewClient(kgoOpts...)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(context.TODO()); err != nil {
		return nil, err
	}

	return &producer{client: client, logger: opts.Logger}, nil
}
