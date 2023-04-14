package main

import (
	"context"
	"fmt"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	serverlessv1 "github.com/kloudlite/operator/apis/serverless/v1"
	"github.com/kloudlite/operator/operator"
	statusWatcher "github.com/kloudlite/operator/operators/status-n-billing/internal/controllers/status-watcher"
	env "github.com/kloudlite/operator/operators/status-n-billing/internal/env"
	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/redpanda"
)

func main() {
	ev := env.GetEnvOrDie()

	producer, err := redpanda.NewProducer(
		ev.KafkaBrokers, redpanda.ProducerOpts{
			SASLAuth: &redpanda.KafkaSASLAuth{
				// SASLMechanism: redpanda.ScramSHA256,
				User:     ev.KafkaSASLUsername,
				Password: ev.KafkaSASLPassword,
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
		panic(fmt.Errorf("failed to ping kafka producer as %s", err.Error()))
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
			Producer: producer,
			// Notifier: types.NewNotifier(ev.ClusterId, producer, ev.KafkaTopicStatusUpdates),
		},
		//&pipelineRunWatcher.Reconciler{
		//	Name:       "pipeline-run",
		//	Env:        ev,
		//	Producer:   producer,
		//	KafkaTopic: ev.KafkaTopicPipelineRunUpdates,
		//},
		//&billingWatcher.Reconciler{
		//	Name:     "billing-watcher",
		//	Env:      ev,
		//	Notifier: types.NewNotifier(ev.ClusterId, producer, ev.KafkaTopicBillingUpdates),
		//},
	)
	mgr.Start()
}
