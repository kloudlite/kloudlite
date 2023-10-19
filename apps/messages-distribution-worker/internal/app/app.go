package app

import (
	"context"
	"kloudlite.io/apps/messages-distribution-worker/internal/env"
	"kloudlite.io/pkg/kafka"

	"go.uber.org/fx"
	"kloudlite.io/pkg/logging"
)

type KafkaConn kafka.Conn

type (
	WaitQueueConsumer   kafka.Consumer
	MessagesDistributor kafka.Producer
)

var Module = fx.Module(
	"app",

	fx.Provide(func(conn KafkaConn, ev *env.Env, logger logging.Logger) (WaitQueueConsumer, error) {
		return kafka.NewConsumer(conn, ev.WaitQueueKafkaConsumerGroup, []string{ev.WaitQueueKafkaTopic}, kafka.ConsumerOpts{
			Logger: logger,
		})
	}),
	fx.Invoke(func(lf fx.Lifecycle, consumer WaitQueueConsumer) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return consumer.LifecycleOnStart(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return consumer.LifecycleOnStop(ctx)
			},
		})
	}),

	fx.Provide(func(conn KafkaConn, logger logging.Logger) (MessagesDistributor, error) {
		return kafka.NewProducer(conn, kafka.ProducerOpts{Logger: logger})
	}),
	fx.Invoke(func(lf fx.Lifecycle, producer MessagesDistributor) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return producer.LifecycleOnStart(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return producer.LifecycleOnStop(ctx)
			},
		})
	}),

	fx.Provide(func(consumer WaitQueueConsumer, producer MessagesDistributor, ev *env.Env, logger logging.Logger) *DistributorClient {
		return NewDistributor(consumer, producer, ev, logger)
	}),

	fx.Invoke(func(lf fx.Lifecycle, distributor *DistributorClient) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go distributor.StartDistributing()
				return nil
			},
			OnStop: func(context.Context) error {
				distributor.StopDistributing()
				return nil
			},
		})
	}),
)
