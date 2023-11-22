package watch_and_update

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/kloudlite/operator/operators/resource-watcher/internal/env"
	t "github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/logging"
	"github.com/kloudlite/operator/pkg/redpanda"
)

type kafkaMsgSender struct {
	errCh                chan error
	kp                   redpanda.Producer
	resourceUpdatesTopic string
	infraUpdatesTopic    string
	logger               logging.Logger
}

// DispatchInfraUpdates implements MessageSender.
func (k *kafkaMsgSender) DispatchInfraUpdates(ctx context.Context, ru t.ResourceUpdate) error {
	b, err := json.Marshal(ru)
	if err != nil {
		return err
	}

	tctx, cf := context.WithTimeout(ctx, 2*time.Second)
	defer cf()

	if _, err := k.kp.Produce(tctx, k.infraUpdatesTopic, ru.ClusterName, b); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			k.logger.Infof("conxtext deadline exceeded, took more than 2 seconds")
		}
		k.errCh <- err
		return err
	}

	k.logger.WithKV("timestamp", time.Now()).Infof("dispatched update to kakfa topic: %s", k.infraUpdatesTopic)
	return nil
}

// DispatchResourceUpdates implements MessageSender.
func (k *kafkaMsgSender) DispatchResourceUpdates(ctx context.Context, ru t.ResourceUpdate) error {
	b, err := json.Marshal(ru)
	if err != nil {
		return err
	}

	tctx, cf := context.WithTimeout(ctx, 2*time.Second)
	defer cf()

	if _, err := k.kp.Produce(tctx, k.resourceUpdatesTopic, ru.ClusterName, b); err != nil {
		if err == context.DeadlineExceeded {
			k.logger.Infof("conxtext deadline exceeded, took more than 2 seconds")
			return nil
		}
		k.errCh <- err
		return err
	}

	k.logger.WithKV("timestamp", time.Now()).Infof("dispatched update to kakfa topic: %s", k.resourceUpdatesTopic)
	return nil
}

func NewKafkaMessageSender(ctx context.Context, ev *env.Env, logger logging.Logger) (MessageSender, error) {
	kp, err := redpanda.NewProducer(ev.KafkaBrokers, redpanda.ProducerOpts{
		Logger:   logger,
		SASLAuth: nil,
	})
	if err != nil {
		return nil, err
	}

	if err := kp.Ping(ctx); err != nil {
		return nil, err
	}

	return &kafkaMsgSender{
		kp:                   kp,
		logger:               logger,
		resourceUpdatesTopic: ev.KafkaResourceUpdatesTopic,
		infraUpdatesTopic:    ev.KafkaInfraUpdatesTopic,
	}, nil
}
