package main

import (
	"context"
	"time"

	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	mongodbMsvcv1 "operators.kloudlite.io/apis/mongodb.msvc/v1"
	mysqlMsvcv1 "operators.kloudlite.io/apis/mysql.msvc/v1"
	redisMsvcv1 "operators.kloudlite.io/apis/redis.msvc/v1"
	serverlessv1 "operators.kloudlite.io/apis/serverless/v1"
	"operators.kloudlite.io/pkg/errors"
	"operators.kloudlite.io/pkg/redpanda"
	"operators.kloudlite.io/operator"
	billingWatcher "operators.kloudlite.io/operators/status-n-billing/internal/controllers/billing-watcher"
	pipelineRunWatcher "operators.kloudlite.io/operators/status-n-billing/internal/controllers/pipeline-run-watcher"
	statusWatcher "operators.kloudlite.io/operators/status-n-billing/internal/controllers/status-watcher"
	env "operators.kloudlite.io/operators/status-n-billing/internal/env"
	"operators.kloudlite.io/operators/status-n-billing/internal/types"
)

func main() {
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

	timeout, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()
	if err := producer.Ping(timeout); err != nil {
		panic("failed to ping kafka producer")
	}

	mgr := operator.New("status-n-billing-watcher")
	mgr.AddToSchemes(
		crdsv1.AddToScheme,
		mongodbMsvcv1.AddToScheme, mysqlMsvcv1.AddToScheme, redisMsvcv1.AddToScheme,
		serverlessv1.AddToScheme,
	)
	mgr.RegisterControllers(
		&statusWatcher.Reconciler{
			Name:     "status-watcher",
			Env:      ev,
			Notifier: types.NewNotifier(ev.ClusterId, producer, ev.KafkaTopicStatusUpdates),
		},
		&pipelineRunWatcher.Reconciler{
			Name:       "pipeline-run",
			Env:        ev,
			Producer:   producer,
			KafkaTopic: ev.KafkaTopicPipelineRunUpdates,
		},
		&billingWatcher.Reconciler{
			Name:     "billing-watcher",
			Env:      ev,
			Notifier: types.NewNotifier(ev.ClusterId, producer, ev.KafkaTopicBillingUpdates),
		},
	)
	mgr.Start()
}
