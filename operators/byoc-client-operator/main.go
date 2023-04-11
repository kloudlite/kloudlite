package main

import (
	"context"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	redpandaMsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/byoc-client-operator/internal/controller"
	"github.com/kloudlite/operator/operators/byoc-client-operator/internal/env"
	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/redpanda"
)

func main() {
	ev, err := env.LoadEnv()
	if err != nil {
		panic(err)
	}

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

	timeout, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()
	if err := producer.Ping(timeout); err != nil {
		panic("failed to ping kafka producer")
	}

	mgr := operator.New("byoc-client")
	mgr.AddToSchemes(clustersv1.AddToScheme, redpandaMsvcv1.AddToScheme)
	mgr.RegisterControllers(&controller.Reconciler{
		Name:     "byoc-client",
		Producer: producer,
		Env:      ev,
	})
	mgr.Start()
}
