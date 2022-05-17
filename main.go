package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/redhat-cop/operator-utils/pkg/util"

	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/controllers/crds"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"operators.kloudlite.io/agent"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"

	"go.uber.org/fx"

	mongodbCluster "operators.kloudlite.io/apis/mongodb-cluster.msvc/v1"
	mongodbStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	crdsControllers "operators.kloudlite.io/controllers/crds"
	mongodbClusterControllers "operators.kloudlite.io/controllers/mongodb-cluster.msvc"
	mongodbStandaloneControllers "operators.kloudlite.io/controllers/mongodb-standalone.msvc"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(crdsv1.AddToScheme(scheme))
	utilruntime.Must(mongodbStandalone.AddToScheme(scheme))
	utilruntime.Must(mongodbCluster.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func fromEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Errorf("ENV '%v' is not provided", key))
	}
	return value
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":9091", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":9092", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                     scheme,
		MetricsBindAddress:         metricsAddr,
		Port:                       9443,
		HealthProbeBindAddress:     probeAddr,
		LeaderElection:             enableLeaderElection,
		LeaderElectionID:           "bf38d2f9.kloudlite.io",
		LeaderElectionResourceLock: "configmaps",
	})

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	clientset := kubernetes.NewForConfigOrDie(mgr.GetConfig())

	harborUserName := fromEnv("HARBOR_USERNAME")
	harborPassword := fromEnv("HARBOR_PASSWORD")

	kafkaBrokers := fromEnv("KAFKA_BROKERS")
	kafkaReplyTopic := fromEnv("KAFKA_REPLY_TOPIC")

	agentKafkaGroupId := fromEnv("AGENT_KAFKA_GROUP_ID")
	agentKafkaTopic := fromEnv("AGENT_KAFKA_TOPIC")

	kafkaProducer, err := kafka.NewProducer(
		&kafka.ConfigMap{
			"bootstrap.servers": kafkaBrokers,
		},
	)

	if err != nil {
		panic(errors.NewEf(err, "could not create kafka producer"))
	}

	fmt.Println("kafka producer connected")

	sender := NewMsgSender(kafkaProducer, kafkaReplyTopic)

	if err = (&crds.ProjectReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		ClientSet:      clientset,
		MessageSender:  sender,
		HarborUserName: harborUserName,
		HarborPassword: harborPassword,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Project")
		os.Exit(1)
	}

	if err = (&crds.AppReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		ClientSet:      clientset,
		MessageSender:  sender,
		HarborUserName: harborUserName,
		HarborPassword: harborPassword,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "App")
		os.Exit(1)
	}
	//
	if err = (&crds.RouterReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		MessageSender: sender,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Router")
		os.Exit(1)
	}

	if err = (&crds.ManagedServiceReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		ClientSet:     clientset,
		MessageSender: sender,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ManagedService")
		os.Exit(1)
	}

	if err = (&crds.ManagedResourceReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		MessageSender: sender,
		ClientSet:     clientset,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ManagedResource")
		os.Exit(1)
	}

	if err = (&crdsControllers.AccountReconciler{
		ReconcilerBase: util.NewReconcilerBase(
			mgr.GetClient(),
			mgr.GetScheme(),
			mgr.GetConfig(),
			mgr.GetEventRecorderFor("Account_controller"),
			mgr.GetAPIReader(),
		),
		Log: ctrl.Log.WithName("controllers").WithName("Account"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Account")
		os.Exit(1)
	}
	if err = (&mongodbStandaloneControllers.ServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Service")
		os.Exit(1)
	}
	if err = (&mongodbStandaloneControllers.DatabaseReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Database")
		os.Exit(1)
	}
	if err = (&mongodbClusterControllers.DatabaseReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Database")
		os.Exit(1)
	}
	if err = (&mongodbClusterControllers.ServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Service")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")

	app := fx.New(
		agent.App(),

		fx.Provide(func() *kafka.Producer {
			return kafkaProducer
		}),

		fx.Provide(func() *kafka.Consumer {
			c, e := kafka.NewConsumer(&kafka.ConfigMap{
				"bootstrap.servers":  kafkaBrokers,
				"group.id":           agentKafkaGroupId,
				"auto.offset.reset":  "earliest",
				"enable.auto.commit": "false",
			})
			if e != nil {
				panic(errors.NewEf(err, "could not create kafka consumer"))
			}
			return c
		}),
		fx.Invoke(func(lf fx.Lifecycle, k *kafka.Consumer) {
			lf.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return k.Subscribe(agentKafkaTopic, nil)
				},
			})
		}),
		fx.Invoke(func(lf fx.Lifecycle) {
			lf.Append(fx.Hook{
				OnStart: func(context.Context) error {
					go func() {
						if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
							setupLog.Error(err, "problem running manager")
							panic(err)
						}
					}()
					return nil
				},
			})
		}),
	)

	app.Run()
}

type msgSender struct {
	kp     *kafka.Producer
	ktopic *string
}

func (m *msgSender) SendMessage(key string, msg lib.MessageReply) error {
	msgBody, e := json.Marshal(msg)
	if e != nil {
		fmt.Println(e)
		return e
	}
	return m.kp.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic: m.ktopic,
		},
		Key:   []byte(key),
		Value: msgBody,
	}, nil)
}

func NewMsgSender(kp *kafka.Producer, ktopic string) lib.MessageSender {
	return &msgSender{
		kp,
		&ktopic,
	}
}
