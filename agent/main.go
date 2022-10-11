package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/codingconcepts/env"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/logging"
	"operators.kloudlite.io/lib/redpanda"
	"sigs.k8s.io/yaml"
)

type AgentMessage struct {
	Action  string         `json:"action"`
	Payload map[string]any `json:"payload,omitempty"`
	Yamls   []byte         `json:"yamls,omitempty"`
}

type ErrMessage struct {
	Error  string `json:"error"`
	Action string `json:"action"`
	// Payload map[string]any `json:"payload"`
	Payload []byte `json:"payload"`
}

type Env struct {
	KafkaSASLUser      string                 `env:"KAFKA_SASL_USER" required:"true"`
	KafkaSASLPassword  string                 `env:"KAFKA_SASL_PASSWORD" required:"true"`
	KafkaSASLMechanism redpanda.SASLMechanism `env:"KAFKA_SASL_MECHANISM"`

	KafkaBrokers           string `env:"KAFKA_BROKERS" required:"true"`
	KafkaConsumerGroupId   string `env:"KAFKA_CONSUMER_GROUP_ID" required:"true"`
	KafkaIncomingTopic     string `env:"KAFKA_INCOMING_TOPIC" required:"true"`
	KafkaErrorOnApplyTopic string `env:"KAFKA_ERROR_ON_APPLY_TOPIC" required:"true"`
}

func main() {
	var dev bool
	flag.BoolVar(&dev, "dev", false, "--dev")
	flag.Parse()

	logr := logging.NewOrDie(&logging.Options{Name: "kloudlite-agent", Dev: dev})

	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}

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

	errProducer, err := redpanda.NewProducer(
		ev.KafkaBrokers, redpanda.ProducerOpts{
			SASLAuth: &kafkaSasl,
		},
	)
	if err != nil {
		panic(err)
	}

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
██   ██ ███████ ██   ██ ██████     ██    
	`,
	)
	logr.Infof("ready for consuming messages")

	consumer.StartConsuming(
		func(kMsg redpanda.KafkaMessage) error {
			logger := logr.WithKV("offset", kMsg.Offset).WithKV("topic", kMsg.Topic).WithKV("partition", kMsg.Partition)
			logger.Infof("received message")
			defer func() {
				logger.Infof("processed message")
			}()

			tctx, cancelFunc := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancelFunc()
			go func() {
				select {
				case <-tctx.Done():
					cancelFunc()
				}
			}()

			var msg AgentMessage
			if err := json.Unmarshal(kMsg.Value, &msg); err != nil {
				logger.Errorf(err, "error when unmarshalling []byte to kafkaMessage : %s", kMsg.Value)
				return err
			}
			switch msg.Action {
			case "apply", "delete", "create":
				{
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

					if errX := func() error {
						c := exec.Command("kubectl", msg.Action, "-f", "-")
						if err != nil {
							return err
						}

						c.Stdin = bytes.NewBuffer(yamls)
						buffOut, buffErr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
						c.Stdout = buffOut
						c.Stderr = buffErr
						if err := c.Run(); err != nil {
							logr.Errorf(err, buffErr.String())
							return err
						}
						return nil
					}(); errX != nil {
						// logger.Infof("failed for action=%s, payload=%s, yamls=%s\n", msg.Action, msg.Payload, msg.Yamls)
						logger.Infof("error: %s", errX.Error())
						errMsg := ErrMessage{
							Action:  msg.Action,
							Error:   errX.Error(),
							Payload: yamls,
						}
						b, err := json.Marshal(errMsg)
						if err != nil {
							return err
						}
						output, err := errProducer.Produce(context.TODO(), ev.KafkaErrorOnApplyTopic, string(kMsg.Key), b)
						if err != nil {
							return err
						}

						logger.Infof(
							"error message published to (topic=%s)", output.Topic,
						)
						return errX
					}
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
