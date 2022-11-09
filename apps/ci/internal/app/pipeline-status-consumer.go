package app

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"kloudlite.io/apps/ci/internal/domain"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
)

type PipelineStatusConsumer redpanda.Consumer

func fxInvokeProcessPipelineRunEvents(d domain.Domain, consumer PipelineStatusConsumer, logr logging.Logger) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	if err := consumer.Ping(ctx); err != nil {
		log.Fatal("failed to ping kafka brokers, for pipeline status consumer")
	}
	logr.Infof("successful ping to kafka brokers")
	consumer.StartConsuming(
		func(msg []byte, _ time.Time, offset int64) error {
			logger := logr.WithName("ci-pipeline-watcher").WithKV("offset", offset)
			logger.Infof("started processing")
			defer func() {
				logger.Infof("processed message")
			}()

			var kMsg domain.PipelineRunStatus
			if err := json.Unmarshal(msg, &kMsg); err != nil {
				return err
			}

			ctx, cancelFn := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancelFn()

			return d.UpdatePipelineRunStatus(ctx, kMsg)
		},
	)
}
