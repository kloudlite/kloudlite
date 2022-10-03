package main

import (
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/redpanda"
	"operators.kloudlite.io/operator"
	billingWatcher "operators.kloudlite.io/operators/status-n-billing/internal/controllers/billing-watcher"
	statusWatcher "operators.kloudlite.io/operators/status-n-billing/internal/controllers/status-watcher"
	"operators.kloudlite.io/operators/status-n-billing/internal/types"
)

func main() {
	mgr := operator.New("status-n-billing-watcher")

	producer, err := redpanda.NewProducer(mgr.Env.KafkaBrokers)
	if err != nil {
		panic(errors.NewEf(err, "creating redpanda producer"))
	}
	defer producer.Close()
	mgr.RegisterControllers(
		&statusWatcher.Reconciler{
			Name:     "status-watcher",
			Notifier: types.NewNotifier(mgr.Env.ClusterId, producer, mgr.Env.KafkaStatusReplyTopic),
		},
		&billingWatcher.Reconciler{
			Name:     "billing-watcher",
			Notifier: types.NewNotifier(mgr.Env.ClusterId, producer, mgr.Env.KafkaBillingReplyTopic),
		},
	)
	mgr.Start()
}
