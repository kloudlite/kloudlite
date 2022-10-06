package main

import (
	"context"
	"time"

	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/redpanda"
	"operators.kloudlite.io/operator"
	billingWatcher "operators.kloudlite.io/operators/status-n-billing/internal/controllers/billing-watcher"
	statusWatcher "operators.kloudlite.io/operators/status-n-billing/internal/controllers/status-watcher"
	env "operators.kloudlite.io/operators/status-n-billing/internal/env"
	"operators.kloudlite.io/operators/status-n-billing/internal/types"
)

func main() {
	mgr := operator.New("status-n-billing-watcher")
	ev := env.GetEnvOrDie()

	producer, err := redpanda.NewProducer(
		ev.KafkaBrokers, redpanda.ProducerOpts{
			SASLAuth: &redpanda.KafkaSASLAuth{
				SASLMechanism: redpanda.ScramSHA256,
				User:          ev.KafkaSASLUsername,
				Password:      ev.KafkaSASLPassword,
			},
		},
	)
	if err != nil {
		panic(errors.NewEf(err, "creating redpanda producer"))
	}
	defer producer.Close()

	timeout, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := producer.Ping(timeout); err != nil {
		panic("failed to ping kafka producer")
	}

	mgr.RegisterControllers(
		&statusWatcher.Reconciler{
			Name:     "status-watcher",
			Env:      ev,
			Notifier: types.NewNotifier(ev.ClusterId, producer, ev.KafkaStatusReplyTopic),
		},
		&billingWatcher.Reconciler{
			Name:     "billing-watcher",
			Env:      ev,
			Notifier: types.NewNotifier(ev.ClusterId, producer, ev.KafkaBillingReplyTopic),
		},
	)
	mgr.Start()
}
