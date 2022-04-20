package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

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

	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/controllers"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(crdsv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8089", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
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
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "bf38d2f9.kloudlite.io",
	})

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	clientset := kubernetes.NewForConfigOrDie(mgr.GetConfig())

	userName, ok := os.LookupEnv("HARBOR_USERNAME")
	if !ok {
		panic(fmt.Errorf("ENV 'HARBOR_USERNAME' is not provided"))
	}

	password, ok := os.LookupEnv("HARBOR_PASSWORD")
	if !ok {
		panic(fmt.Errorf("ENV 'HARBOR_PASSWORD' is not provided"))
	}

	kafkaReplyTopic, ok := os.LookupEnv("KAFKA_REPLY_TOPIC")
	if !ok {
		panic(fmt.Errorf("ENV 'KAFKA_REPLY_TOPIC' is not provided"))
	}

	kafkaBrokers, ok := os.LookupEnv("KAFKA_BROKERS")
	if !ok {
		panic(fmt.Errorf("ENV 'KAFKA_BROKERS' is not provided"))
	}

	kafkaProducer, err := kafka.NewProducer(
		&kafka.ConfigMap{
			"bootstrap.servers": kafkaBrokers,
		},
	)

	if err != nil {
		panic(errors.NewEf(err, "could not create kafka producer"))
	}

	fmt.Println("kafka producer connected")

	sendMessage := func(key string, message lib.MessageReply) error {
		msgBody, e := json.Marshal(message)
		if e != nil {
			fmt.Println(e)
			return e
		}
		return kafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &kafkaReplyTopic,
				Partition: int32(kafka.PartitionAny),
			},
			Key:   []byte(key),
			Value: msgBody,
		}, nil)
	}

	if err = (&controllers.ProjectReconciler{
		Client:      mgr.GetClient(),
		Scheme:      mgr.GetScheme(),
		ClientSet:   clientset,
		SendMessage: sendMessage,
		JobMgr:      lib.NewJobber(clientset),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Project")
		os.Exit(1)
	}

	if err = (&controllers.AppReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		ClientSet:      clientset,
		JobMgr:         lib.NewJobber(clientset),
		SendMessage:    sendMessage,
		HarborUserName: userName,
		HarborPassword: password,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "App")
		os.Exit(1)
	}

	// if err = (&controllers.RouterReconciler{
	// 	Client: mgr.GetClient(),
	// 	Scheme: mgr.GetScheme(),
	// }).SetupWithManager(mgr); err != nil {
	// 	setupLog.Error(err, "unable to create controller", "controller", "Router")
	// 	os.Exit(1)
	// }

	// if err = (&controllers.ManagedServiceReconciler{
	// 	Client:    mgr.GetClient(),
	// 	Scheme:    mgr.GetScheme(),
	// 	ClientSet: clientset,
	// 	JobMgr:    lib.NewJobber(clientset),
	// }).SetupWithManager(mgr); err != nil {
	// 	setupLog.Error(err, "unable to create controller", "controller", "ManagedService")
	// 	os.Exit(1)
	// }

	// if err = (&controllers.ManagedResourceReconciler{
	// 	Client:    mgr.GetClient(),
	// 	Scheme:    mgr.GetScheme(),
	// 	ClientSet: clientset,
	// 	JobMgr:    lib.NewJobber(clientset),
	// }).SetupWithManager(mgr); err != nil {
	// 	setupLog.Error(err, "unable to create controller", "controller", "ManagedResource")
	// 	os.Exit(1)
	// }

	// if err = (&controllers.PipelineReconciler{
	// 	Client: mgr.GetClient(),
	// 	Scheme: mgr.GetScheme(),
	// }).SetupWithManager(mgr); err != nil {
	// 	setupLog.Error(err, "unable to create controller", "controller", "Pipeline")
	// 	os.Exit(1)
	// }
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
