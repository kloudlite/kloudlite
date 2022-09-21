package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	serverlessv1 "operators.kloudlite.io/apis/serverless/v1"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/logging"
	"operators.kloudlite.io/lib/redpanda"
	"operators.kloudlite.io/operators/webhook-worker/internal/env"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type HttpHook struct {
	Body        []byte            `json:"body"`
	Headers     map[string]string `json:"headers"`
	Url         string            `json:"url"`
	QueryParams string            `json:"queryParams"`
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(crdsv1.AddToScheme(scheme))
	utilruntime.Must(serverlessv1.AddToScheme(scheme))
}

func restartApp(kClient client.Client, imageName string) error {
	var apps crdsv1.AppList
	if err := kClient.List(
		context.TODO(), &apps, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(
				map[string]string{
					fmt.Sprintf("kloudlite.io/image-%s", fn.Sha1Sum([]byte(imageName))): "true",
				},
			),
		},
	); err != nil {
		return err
	}

	for _, item := range apps.Items {
		if _, err := controllerutil.CreateOrUpdate(
			context.TODO(), kClient, &item, func() error {
				ann := item.GetAnnotations()
				ann[constants.AnnotationKeys.Restart] = "true"
				item.SetAnnotations(ann)
				return nil
			},
		); err != nil {
			return err
		}
	}
	return nil
}

func restartLambda(kClient client.Client, imageName string) error {
	var lambdaList serverlessv1.LambdaList
	if err := kClient.List(
		context.TODO(), &lambdaList, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(
				map[string]string{
					fmt.Sprintf("kloudlite.io/image-%s", fn.Sha1Sum([]byte(imageName))): "true",
				},
			),
		},
	); err != nil {
		return err
	}

	for _, item := range lambdaList.Items {
		if _, err := controllerutil.CreateOrUpdate(
			context.TODO(), kClient, &item, func() error {
				ann := item.GetAnnotations()
				ann[constants.AnnotationKeys.Restart] = "true"
				item.SetAnnotations(ann)
				return nil
			},
		); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var isDev bool
	var devServerAddr string
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.StringVar(&devServerAddr, "dev-server-addr", "localhost:8080", "--dev-server-addr <host:port>")
	flag.Parse()

	logger := logging.NewOrDie(
		&logging.Options{Name: "webhook-worker", Dev: isDev},
	)

	ev := env.GetEnvOrDie()
	consumer, err := redpanda.NewConsumer(
		ev.KafkaBrokers, ev.KafkaConsumerGroupId, ev.KafkaHarborWebhookIncomingTopic, &redpanda.ConsumerOpts{
			Logger: logger,
		},
	)
	if err != nil {
		panic(err)
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelCtx()

	if err := consumer.Ping(ctx); err != nil {
		panic(err)
	}

	logger.Infof("consumer connected to kafka brokers (successful ping)")

	fmt.Println(
		`
██████  ███████  █████  ██████  ██    ██ 
██   ██ ██      ██   ██ ██   ██  ██  ██  
██████  █████   ███████ ██   ██   ████   
██   ██ ██      ██   ██ ██   ██    ██    
██   ██ ███████ ██   ██ ██████     ██    
	`,
	)

	kClient, err := client.New(
		func() *rest.Config {
			if isDev {
				return &rest.Config{Host: devServerAddr}
			}
			config, err := rest.InClusterConfig()
			if err != nil {
				panic(err)
			}
			return config
		}(), client.Options{},
	)

	s := kClient.Scheme()
	scheme = runtime.NewScheme()
	utilruntime.Must(crdsv1.AddToScheme(scheme))
	utilruntime.Must(serverlessv1.AddToScheme(scheme))
	*s = *scheme

	if err != nil {
		panic(err)
	}

	if err := consumer.StartConsuming(
		func(msg redpanda.KafkaMessage) error {
			log := logger.WithKV("offset", msg.Offset).WithKV("topic", msg.Topic).WithKV("partition", msg.Partition)
			log.Infof("received message")
			var httpHook HttpHook
			if err := json.Unmarshal(msg.Value, &httpHook); err != nil {
				return err
			}

			var harborHookBody harbor.WebhookBody
			if err := json.Unmarshal(httpHook.Body, &harborHookBody); err != nil {
				return err
			}

			imageName := func() string {
				for _, v := range harborHookBody.EventData.Resources {
					if v.ResourceUrl != "" {
						return v.ResourceUrl
					}
				}
				return ""
			}()

			if err := restartApp(kClient, imageName); err != nil {
				return errors.NewEf(err, "restarting apps")
			}

			if err := restartLambda(kClient, imageName); err != nil {
				return errors.NewEf(err, "restarting lambda")
			}

			log.Infof("processed message")
			return nil
		},
	); err != nil {
		panic(err)
	}

}
