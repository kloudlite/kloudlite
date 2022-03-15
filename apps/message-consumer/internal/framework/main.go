package framework

import (
	// "encoding/json"
	"encoding/json"
	"fmt"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"kloudlite.io/apps/message-consumer/internal/app"
	"kloudlite.io/pkg/errors"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func MakeFramework(cfg *Config) (fm Framework, e error) {
	defer errors.HandleErr(&e)

	consumer, e := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  cfg.KafkaBrokers,
		"group.id":           cfg.ConsumerGroupId,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "true",
	})

	errors.AssertNoError(e, fmt.Errorf("failed to create consumer because %v", e))
	log.Infof("(%v) consumers connected to Kafka brokers (%v)", cfg.ConsumerGroupId, cfg.KafkaBrokers)

	e = consumer.SubscribeTopics([]string{
		fmt.Sprintf("^%s", cfg.TopicPrefix),
	}, nil)

	errors.AssertNoError(e, fmt.Errorf("failed to subscribe to topics (%v) because %v", cfg.TopicPrefix, e))
	logrus.Infof("subscribed to topics %v", cfg.TopicPrefix)

	kApplier, e := MakeKubeApplier(cfg.IsDev)
	errors.AssertNoError(e, fmt.Errorf("failed to create k8sApplier because %v", e))

	appSvc := app.MakeApp(kApplier, MakeGqlClient())

	fm = func() {
		if cfg.IsDev {
			err := appSvc.Handle(&app.Message{
				Action:       "create",
				ResourceType: app.ResourceManagedRes,
				ResourceId:   "mres-2zwm7jxrg-0mjmsmnphnlzwdfpfebabs-kl",
			})

			if err != nil {
				log.Errorf("could not handle message because %v", err)
				return
			}
		}
		for {
			fmt.Println("awaiting for new message ...")
			msg, err := consumer.ReadMessage(-1)
			if err != nil {
				log.Errorf("could not read message from kafka because %v", err)
				continue
			}
			log.Infof("received message (topic=%v), %v", msg.TopicPartition.Topic, string(msg.Value))

			var msgData app.Message
			err = json.Unmarshal(msg.Value, &msgData)
			if err != nil {
				log.Errorf("could not unmarshal message because %v", err)
				continue
			}

			err = appSvc.Handle(&msgData)
			if err != nil {
				log.Errorf("could not handle message because %v", err)
				continue
			}
		}
	}

	return
}
