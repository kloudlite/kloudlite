package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kloudlite/operator/agent/internal/env"
	t "github.com/kloudlite/operator/agent/types"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	"github.com/kloudlite/operator/pkg/redpanda"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	logr := logging.NewOrDie(&logging.Options{Name: "kloudlite-agent", Dev: isDev})

	ev := env.GetEnvOrDie()

	yamlClient := func() *kubectl.YAMLClient {
		if isDev {
			return kubectl.NewYAMLClientOrDie(&rest.Config{Host: "localhost:8080"})
		}
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		return kubectl.NewYAMLClientOrDie(config)
	}()

	kafkaSasl := redpanda.KafkaSASLAuth{
		User:          ev.KafkaSASLUser,
		Password:      ev.KafkaSASLPassword,
		SASLMechanism: redpanda.ScramSHA256,
	}

	errProducer, err := redpanda.NewProducer(ev.KafkaBrokers, redpanda.ProducerOpts{SASLAuth: &kafkaSasl})
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

	tctx, cancelFunc := func() (context.Context, context.CancelFunc) {
		if isDev {
			return context.WithCancel(context.TODO())
		}
		return context.WithTimeout(context.TODO(), 5*time.Second)
	}()
	defer cancelFunc()

	if err := consumer.Ping(tctx); err != nil {
		log.Fatal("failed to ping kafka brokers")
	}
	logr.Infof("successful ping to kafka brokers")

	fmt.Println(
		`
	███████  ███████  █████  ██████  ██    ██ 
  ██   ██  ██      ██   ██ ██   ██  ██  ██  
  ██████   █████   ███████ ██   ██   ████   
  ██   ██  ██      ██   ██ ██   ██    ██    
  ██   ██  ███████ ██   ██ ██████     ██      for consuming kafka messages`,
	)

	dispatchErrorButCommit := func(err error, msg t.AgentMessage, logger logging.Logger) error {
		logger.Debugf("[ERROR]: %s", err.Error())
		b, err := json.Marshal(t.AgentErrMessage{
			AccountName: msg.AccountName,
			ClusterName: msg.ClusterName,
			Error:       err,
			Action:      msg.Action,
			Object:      msg.Object,
		})
		if err != nil {
			return err
		}

		obj := unstructured.Unstructured{Object: msg.Object}

		out, err := errProducer.Produce(tctx, ev.KafkaErrorOnApplyTopic, obj.GetNamespace(), b)
		logger.Infof("dispatched error message to topic(%s)", out.Topic)
		return err
	}

	counter := 0

	consumer.StartConsuming(
		func(kMsg redpanda.KafkaMessage) error {
			logger := logr.WithKV("topic", kMsg.Topic)
			counter += 1
			logger.Debugf("====> received kafka message [%d]", counter)
			defer func() {
				logger.Debugf("<==== processed kafka message [%d]", counter)
			}()

			var msg t.AgentMessage
			if err := json.Unmarshal(kMsg.Value, &msg); err != nil {
				logger.Errorf(err, "error when unmarshalling []byte to kafkaMessage : %s", kMsg.Value)
				return nil
			}

			if msg.Object == nil {
				logger.Infof("msg.Object is nil, could not process anything out of this kafka message, ignoring ...")
				return nil
			}

			obj := unstructured.Unstructured{Object: msg.Object}

			logger = logger.WithKV("gvk", obj.GetObjectKind().GroupVersionKind().String()).WithKV("clusterName", msg.ClusterName).WithKV("accountName", msg.AccountName).WithKV("action", msg.Action)

			logger.Infof("received message [%d]", counter)
			defer func() {
				logger.Infof("processed message [%d]", counter)
			}()

			if len(strings.TrimSpace(msg.AccountName)) == 0 {
				return dispatchErrorButCommit(fmt.Errorf("field 'accountName' must be defined in message"), msg, logger)
			}

			tctx, cancelFunc := func() (context.Context, context.CancelFunc) {
				if isDev {
					return context.WithCancel(context.TODO())
				}
				return context.WithTimeout(context.TODO(), 3*time.Second)
			}()

			defer cancelFunc()
			go func() {
				<-tctx.Done()
				cancelFunc()
			}()

			switch msg.Action {
			case "apply", "delete", "create":
				{
					b, err := yaml.Marshal(msg.Object)
					if err != nil {
						return dispatchErrorButCommit(err, msg, logger)
					}

					if msg.Action == "apply" {
						_, err := yamlClient.ApplyYAML(tctx, b)
						if err != nil {
							return dispatchErrorButCommit(err, msg, logger)
						}
						return nil
					}

					if msg.Action == "delete" {
						err := yamlClient.DeleteYAML(tctx, b)
						if err != nil {
							return dispatchErrorButCommit(err, msg, logger)
						}
						return nil
					}
					return nil
				}
			default:
				{
					return dispatchErrorButCommit(fmt.Errorf("invalid action (%s)", msg.Action), msg, logger)
				}
			}
		},
	)
}
