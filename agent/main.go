package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/kloudlite/operator/agent/internal/env"
	t "github.com/kloudlite/operator/agent/internal/types"
	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	"github.com/kloudlite/operator/pkg/redpanda"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

func main() {
	var dev bool
	flag.BoolVar(&dev, "dev", false, "--dev")
	flag.Parse()

	logr := logging.NewOrDie(&logging.Options{Name: "kloudlite-agent", Dev: dev})

	ev := env.GetEnvOrDie()

	yamlClient := func() *kubectl.YAMLClient {
		if dev {
			return kubectl.NewYAMLClientOrDie(&rest.Config{Host: "localhost:8080"})
		}
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		return kubectl.NewYAMLClientOrDie(config)
	}()

	kafkaSasl := redpanda.KafkaSASLAuth{
		SASLMechanism: func() redpanda.SASLMechanism {
			if ev.KafkaSASLMechanism != "" {
				return ev.KafkaSASLMechanism
			}
			return redpanda.ScramSHA256
		}(),
		User:     ev.KafkaSASLUser,
		Password: ev.KafkaSASLPassword,
	}

	// errProducer, err := redpanda.NewProducer(
	// 	ev.KafkaBrokers, redpanda.ProducerOpts{
	// 		SASLAuth: &kafkaSasl,
	// 	},
	// )
	// if err != nil {
	// 	panic(err)
	// }

	consumer, err := redpanda.NewConsumer(
		ev.KafkaBrokers, ev.KafkaConsumerGroupId,
		ev.KafkaIncomingTopic, redpanda.ConsumerOpts{Logger: logr, SASLAuth: &kafkaSasl},
	)
	if err != nil {
		panic(err)
	}

	tctx, cancelFunc := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancelFunc()

	if err := consumer.Ping(tctx); err != nil {
		log.Fatal("failed to ping kafka brokers")
	}
	logr.Infof("successful ping to kafka brokers")

	fmt.Println(
		`
██████  ███████  █████  ██████  ██    ██ 
██   ██ ██      ██   ██ ██   ██  ██  ██  
██████  █████   ███████ ██   ██   ████   
██   ██ ██      ██   ██ ██   ██    ██    
██   ██ ███████ ██   ██ ██████     ██`,
	)
	logr.Infof("ready for consuming messages")

	consumer.StartConsuming(
		func(kMsg redpanda.KafkaMessage) error {
			logger := logr.WithKV("offset", kMsg.Offset).WithKV("topic", kMsg.Topic).WithKV("partition", kMsg.Partition)
			logger.Infof("received message")
			defer func() {
				logger.Infof("processed message")
			}()
			tctx, cancelFunc := func() (context.Context, context.CancelFunc) {
				if dev {
					return context.WithCancel(context.TODO())
				}
				return context.WithTimeout(context.TODO(), 3*time.Second)
			}()
			defer cancelFunc()
			go func() {
				select {
				case <-tctx.Done():
					cancelFunc()
				}
			}()

			var msg t.AgentMessage
			if err := json.Unmarshal(kMsg.Value, &msg); err != nil {
				logger.Errorf(err, "error when unmarshalling []byte to kafkaMessage : %s", kMsg.Value)
				return err
			}
			switch msg.Action {
			case "apply", "delete", "create":
				{
					mLogger := logger.WithKV("action", msg.Action)
					mLogger.WithKV("action", msg.Action).Infof("received message")
					defer func() {
						mLogger.WithKV("action", msg.Action).Infof("processed message")
					}()

					yamls, err := func() ([]byte, error) {
						if msg.Yamls != nil {
							return msg.Yamls, nil
						}

						pb, err := json.Marshal(msg.Payload)
						if err != nil {
							return nil, errors.NewEf(err, "could not convert msg.Payload into []byte")
						}
						yb, err := yaml.JSONToYAML(pb)
						if err != nil {
							return nil, errors.NewEf(err, "could not convert JSON to YAML")
						}
						return yb, nil
					}()
					if err != nil {
						return err
					}

					// with-api
					if msg.Action == "apply" {
						return yamlClient.ApplyYAML(tctx, yamls)
					}
					if msg.Action == "delete" {
						return yamlClient.DeleteYAML(tctx, yamls)
					}
					return nil
				}
			default:
				{
					logger.Errorf(nil, "Invalid Action: %s", msg.Action)
				}
			}
			return nil
		},
	)
}
