package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"

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
	KafkaBrokers           string `env:"KAFKA_BROKERS" required:"true"`
	KafkaConsumerGroupId   string `env:"KAFKA_CONSUMER_GROUP_ID" required:"true"`
	KafkaIncomingTopic     string `env:"KAFKA_INCOMING_TOPIC" required:"true"`
	KafkaErrorOnApplyTopic string `env:"KAFKA_ERROR_ON_APPLY_TOPIC" required:"true"`
}

func main() {
	var dev bool
	flag.BoolVar(&dev, "dev", false, "--dev")
	flag.Parse()

	logger := logging.NewOrDie(&logging.Options{Name: "kloudlite-agent", Dev: dev})

	var envVars Env
	if err := env.Set(&envVars); err != nil {
		panic(err)
	}

	errProducer, err := redpanda.NewProducer(envVars.KafkaBrokers)
	if err != nil {
		panic(err)
	}

	consumer, err := redpanda.NewConsumer(
		envVars.KafkaBrokers, envVars.KafkaConsumerGroupId,
		envVars.KafkaIncomingTopic, &redpanda.ConsumerOptions{
			ErrProducer: errProducer,
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(
		`
██████  ███████  █████  ██████  ██    ██ 
██   ██ ██      ██   ██ ██   ██  ██  ██  
██████  █████   ███████ ██   ██   ████   
██   ██ ██      ██   ██ ██   ██    ██    
██   ██ ███████ ██   ██ ██████     ██    
	`,
	)
	logger.Infof("ready for consuming messages")

	consumer.StartConsuming(
		func(kMsg *redpanda.KafkaMessage) error {
			var msg AgentMessage
			if err := json.Unmarshal(kMsg.Value, &msg); err != nil {
				logger.Errorf(err, "error when unmarshalling []byte to kafkaMessage : %s", kMsg.Value)
				return err
			}
			logger.Debugf("action=%s, payload=%s\n", msg.Action, msg.Payload)

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
						c.Stdout = os.Stdout
						errStream := bytes.NewBuffer([]byte{})
						c.Stderr = errStream
						if err := c.Run(); err != nil {
							logger.Errorf(err, errStream.String())
							return errors.NewEf(err, errStream.String())
						}
						return nil
					}(); errX != nil {
						errMsg := ErrMessage{
							Action:  msg.Action,
							Error:   errX.Error(),
							Payload: yamls,
						}
						b, err := json.Marshal(errMsg)
						if err != nil {
							return err
						}
						if err := errProducer.Produce(context.TODO(), envVars.KafkaErrorOnApplyTopic, string(kMsg.Key), b); err != nil {
							return err
						}
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
