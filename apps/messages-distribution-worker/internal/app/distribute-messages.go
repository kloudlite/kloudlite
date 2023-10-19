package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"kloudlite.io/apps/messages-distribution-worker/internal/env"
	"kloudlite.io/pkg/kafka"
	"kloudlite.io/pkg/logging"
)

type DistributorClient struct {
	envVars  *env.Env
	counter  int64
	consumer WaitQueueConsumer
	producer MessagesDistributor
	logger   logging.Logger

	topicsMap map[string]struct{}
}

func (d *DistributorClient) listTopics() error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/topics", d.envVars.RedpandaHttpAddr), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var topics []string
	if err := json.Unmarshal(b, &topics); err != nil {
		return err
	}

	topicsMap := make(map[string]struct{}, len(topics))
	for _, topic := range topics {
		topicsMap[topic] = struct{}{}
	}

	d.topicsMap = topicsMap
	d.logger.Debugf("listing topics, cleared cache: (total: %d)", len(topics))
	return nil
}

func (d *DistributorClient) topicExists(topicName string) bool {
	if d.topicsMap == nil {
		d.logger.Infof("topics map is nil, this should not have happened")
		return false
	}

	if _, ok := d.topicsMap[topicName]; ok {
		return true
	}

	return false
}

func (d *DistributorClient) StartDistributing() {
	d.listTopics()
	go func() {
		for {
			time.Sleep(5 * time.Second)
			d.listTopics()
		}
	}()

	d.consumer.StartConsuming(func(ctx kafka.ConsumerContext, topic string, value []byte, metadata kafka.RecordMetadata) error {
		topicName := string(metadata.Headers["topic"])

		if !d.topicExists(topicName) {
			ctx.Logger.Infof("topic %s does not exist, creating ...", topicName)
			if b, err := exec.Command("rpk", "topic", "create", topicName, "-p", d.envVars.NewTopicPartitionsCount, "-r", d.envVars.NewTopicReplicationCount, "--brokers", d.envVars.KafkaBrokers).CombinedOutput(); err != nil {
				ctx.Logger.Errorf(err, string(b))
				// return err
			}
			d.listTopics() // updating topics cache
		}

		ctx.Logger.Debugf("topic %s exists, about to produce message ...", topicName)
		if _, err := d.producer.Produce(ctx, topicName, value, kafka.MessageArgs{
			Key:     metadata.Key,
			Headers: nil,
		}); err != nil {
			ctx.Logger.Errorf(err, "error while producing message to topic %s", topicName)
			return err
		}
		d.counter += 1
		ctx.Logger.Infof("[%d] mirrored message from (topic: %s) to (topic: %s)", d.counter, topic, topicName)
		return nil
	})
}

func (d *DistributorClient) StopDistributing() {
	d.consumer.StopConsuming()
}

func NewDistributor(consumer WaitQueueConsumer, producer MessagesDistributor, ev *env.Env, logger logging.Logger) *DistributorClient {
	return &DistributorClient{
		counter:   0,
		envVars:   ev,
		consumer:  consumer,
		producer:  producer,
		logger:    logger,
		topicsMap: map[string]struct{}{},
	}
}
