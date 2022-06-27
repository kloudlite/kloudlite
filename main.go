package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/types"

	"operators.kloudlite.io/lib/logger"
	"operators.kloudlite.io/lib/redpanda"
	t "operators.kloudlite.io/lib/types"

	"github.com/redhat-cop/operator-utils/pkg/util"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/controllers/crds"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"operators.kloudlite.io/agent"
	elasticsearchmsvcv1 "operators.kloudlite.io/apis/elasticsearch.msvc/v1"
	influxdbmsvcv1 "operators.kloudlite.io/apis/influxdb.msvc/v1"
	mongodbCluster "operators.kloudlite.io/apis/mongodb-cluster.msvc/v1"
	mongodbStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	mysqlclustermsvcv1 "operators.kloudlite.io/apis/mysql-cluster.msvc/v1"
	mysqlstandalonemsvcv1 "operators.kloudlite.io/apis/mysql-standalone.msvc/v1"
	opensearchmsvcv1 "operators.kloudlite.io/apis/opensearch.msvc/v1"
	redisclustermsvcv1 "operators.kloudlite.io/apis/redis-cluster.msvc/v1"
	redisstandalonemsvcv1 "operators.kloudlite.io/apis/redis-standalone.msvc/v1"
	s3awsv1 "operators.kloudlite.io/apis/s3.aws/v1"
	serverlessv1 "operators.kloudlite.io/apis/serverless/v1"
	elasticsearchmsvccontrollers "operators.kloudlite.io/controllers/elasticsearch.msvc"
	influxdbmsvccontrollers "operators.kloudlite.io/controllers/influxdb.msvc"
	mongodbStandaloneControllers "operators.kloudlite.io/controllers/mongodb-standalone.msvc"
	mysqlStandaloneController "operators.kloudlite.io/controllers/mysql-standalone.msvc"
	opensearchmsvccontrollers "operators.kloudlite.io/controllers/opensearch.msvc"
	redisstandalonemsvccontrollers "operators.kloudlite.io/controllers/redis-standalone.msvc"
	s3awscontrollers "operators.kloudlite.io/controllers/s3.aws"
	serverlesscontrollers "operators.kloudlite.io/controllers/serverless"
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
	utilruntime.Must(mysqlstandalonemsvcv1.AddToScheme(scheme))
	utilruntime.Must(mysqlclustermsvcv1.AddToScheme(scheme))
	utilruntime.Must(redisstandalonemsvcv1.AddToScheme(scheme))
	utilruntime.Must(redisclustermsvcv1.AddToScheme(scheme))
	utilruntime.Must(influxdbmsvcv1.AddToScheme(scheme))
	utilruntime.Must(serverlessv1.AddToScheme(scheme))
	utilruntime.Must(elasticsearchmsvcv1.AddToScheme(scheme))
	utilruntime.Must(opensearchmsvcv1.AddToScheme(scheme))
	utilruntime.Must(s3awsv1.AddToScheme(scheme))
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
	// flag.StringVar(&metricsAddr, "metrics-bind-address", ":9091", "The address the metric endpoint binds to.")
	// flag.StringVar(&probeAddr, "health-probe-bind-address", ":9092", "The address the probe endpoint binds to.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":12345", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":12346", "The address the probe endpoint binds to.")
	flag.BoolVar(
		&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.",
	)

	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "Enable development mode")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := func() (manager.Manager, error) {
		cOpts := ctrl.Options{
			Scheme:                     scheme,
			MetricsBindAddress:         metricsAddr,
			Port:                       9443,
			HealthProbeBindAddress:     probeAddr,
			LeaderElection:             enableLeaderElection,
			LeaderElectionID:           "bf38d2f9.kloudlite.io",
			LeaderElectionResourceLock: "configmaps",
		}
		if isDev {
			return ctrl.NewManager(&rest.Config{Host: "localhost:8080"}, cOpts)
		}
		return ctrl.NewManager(ctrl.GetConfigOrDie(), cOpts)
	}()
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	harborUserName := fromEnv("HARBOR_USERNAME")
	harborPassword := fromEnv("HARBOR_PASSWORD")

	kafkaBrokers := fromEnv("KAFKA_BROKERS")
	kafkaReplyTopic := fromEnv("KAFKA_REPLY_TOPIC")

	agentKafkaGroupId := fromEnv("AGENT_KAFKA_GROUP_ID")
	agentKafkaTopic := fromEnv("AGENT_KAFKA_TOPIC")

	producer, err := redpanda.NewProducer(kafkaBrokers)
	if err != nil {
		setupLog.Error(err, "creating redpanda producer")
		panic(err)
	}
	defer producer.Close()

	messageSender := NewMsgSender(producer, kafkaReplyTopic)

	consumer, err := redpanda.NewConsumer(kafkaBrokers, agentKafkaGroupId, agentKafkaTopic)
	if err != nil {
		setupLog.Error(err, "creating redpanda consumer")
		panic(err)
	}
	consumer.SetupLogger(logger.NewZapLogger(types.NamespacedName{}))
	defer consumer.Close()

	if err = (&crds.ProjectReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		MessageSender:  messageSender,
		HarborUserName: harborUserName,
		HarborPassword: harborPassword,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Project")
		os.Exit(1)
	}

	if err = (&crds.AppReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		MessageSender:  messageSender,
		HarborUserName: harborUserName,
		HarborPassword: harborPassword,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "App")
		os.Exit(1)
	}

	if err = (&crds.RouterReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		MessageSender: messageSender,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Router")
		os.Exit(1)
	}

	if err = (&crds.ManagedServiceReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		MessageSender: messageSender,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ManagedService")
		os.Exit(1)
	}

	if err = (&crds.ManagedResourceReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		MessageSender: messageSender,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ManagedResource")
		os.Exit(1)
	}

	if err = (&crds.AccountReconciler{
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
	//
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

	// if err = (&mongodbClusterControllers.DatabaseReconciler{
	//	Client: mgr.GetClient(),
	//	Scheme: mgr.GetScheme(),
	// }).SetupWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create controller", "controller", "Database")
	//	os.Exit(1)
	// }
	//
	// if err = (&mongodbClusterControllers.ServiceReconciler{
	//	Client: mgr.GetClient(),
	//	Scheme: mgr.GetScheme(),
	// }).SetupWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create controller", "controller", "Service")
	//	os.Exit(1)
	// }

	if err = (&mysqlStandaloneController.ServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Service")
		os.Exit(1)
	}

	if err = (&mysqlStandaloneController.DatabaseReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Database")
		os.Exit(1)
	}

	// if err = (&mysqlclustermsvccontrollers.ServiceReconciler{
	//	Client: mgr.GetClient(),
	//	Scheme: mgr.GetScheme(),
	// }).SetupWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create controller", "controller", "Service")
	//	os.Exit(1)
	// }
	//
	// if err = (&mysqlclustermsvccontrollers.DatabaseReconciler{
	//	Client: mgr.GetClient(),
	//	Scheme: mgr.GetScheme(),
	// }).SetupWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create controller", "controller", "Database")
	//	os.Exit(1)
	// }
	//
	if err = (&redisstandalonemsvccontrollers.ServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Service")
		os.Exit(1)
	}

	if err = (&redisstandalonemsvccontrollers.ACLAccountReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ACLAccount")
		os.Exit(1)
	}

	if err = (&serverlesscontrollers.LambdaReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Lambda")
		os.Exit(1)
	}

	if err = (&elasticsearchmsvccontrollers.ServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Service")
		os.Exit(1)
	}
	if err = (&opensearchmsvccontrollers.ServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Service")
		os.Exit(1)
	}
	if err = (&influxdbmsvccontrollers.ServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Service")
		os.Exit(1)
	}
	if err = (&influxdbmsvccontrollers.BucketReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Bucket")
		os.Exit(1)
	}
	if err = (&s3awscontrollers.BucketReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Bucket")
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

	go agent.Run(consumer)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		panic(err)
	}
}

type msgSender struct {
	producer *redpanda.Producer
	kTopic   string
}

func (m *msgSender) SendMessage(ctx context.Context, key string, message t.MessageReply) error {
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return m.producer.Produce(ctx, m.kTopic, key, b)
}

func NewMsgSender(producer *redpanda.Producer, kTopic string) *msgSender {
	return &msgSender{producer: producer, kTopic: kTopic}
}
