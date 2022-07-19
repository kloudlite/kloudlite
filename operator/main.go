package main

import (
	"flag"
	"os"

	"operators.kloudlite.io/env"

	"k8s.io/apimachinery/pkg/types"

	"github.com/redhat-cop/operator-utils/pkg/util"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"operators.kloudlite.io/lib/logging"
	"operators.kloudlite.io/lib/redpanda"

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
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
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
	artifactscontrollers "operators.kloudlite.io/controllers/artifacts"
	elasticsearchmsvccontrollers "operators.kloudlite.io/controllers/elasticsearch.msvc"
	influxdbmsvccontrollers "operators.kloudlite.io/controllers/influxdb.msvc"
	mongodbStandaloneControllers "operators.kloudlite.io/controllers/mongodb-standalone.msvc"
	mysqlStandaloneController "operators.kloudlite.io/controllers/mysql-standalone.msvc"
	opensearchmsvccontrollers "operators.kloudlite.io/controllers/opensearch.msvc"
	redisstandalonemsvccontrollers "operators.kloudlite.io/controllers/redis-standalone.msvc"
	s3awscontrollers "operators.kloudlite.io/controllers/s3.aws"
	serverlesscontrollers "operators.kloudlite.io/controllers/serverless"
	watchercontrollers "operators.kloudlite.io/controllers/watcher"
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
	utilruntime.Must(artifactsv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
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

	myLogger := logging.NewOrDie(
		&logging.Options{
			Name: "operator-logger",
			Dev:  isDev,
		},
	)

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

	envVars := env.Must(env.GetEnv())

	if err = (&crds.ProjectReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Env:    envVars,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Project")
		os.Exit(1)
	}

	if err = (&crds.AppReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Env:    envVars,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "App")
		os.Exit(1)
	}

	if err = (&crds.RouterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Env:    envVars,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Router")
		os.Exit(1)
	}

	if err = (&crds.ManagedServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ManagedService")
		os.Exit(1)
	}

	if err = (&crds.ManagedResourceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
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
		Env:    envVars,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Bucket")
		os.Exit(1)
	}

	producer, err := redpanda.NewProducer(envVars.KafkaBrokers)
	if err != nil {
		setupLog.Error(err, "creating redpanda producer")
		panic(err)
	}
	defer producer.Close()

	statusNotifier := watchercontrollers.NewNotifier(envVars.ClusterId, producer, envVars.KafkaStatusReplyTopic)
	billingNotifier := watchercontrollers.NewNotifier(envVars.ClusterId, producer, envVars.KafkaBillingReplyTopic)

	if err = (&watchercontrollers.StatusWatcherReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Env:      envVars,
		Notifier: statusNotifier,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "StatusWatcher")
		os.Exit(1)
	}
	if err = (&watchercontrollers.BillingWatcherReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Env:      envVars,
		Notifier: billingNotifier,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BillingWatcher")
		os.Exit(1)
	}

	if err = (&artifactscontrollers.HarborProjectReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Env:    envVars,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HarborProject")
		os.Exit(1)
	}
	if err = (&artifactscontrollers.HarborUserAccountReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Env:    envVars,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HarborUserAccount")
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

	consumer, err := redpanda.NewConsumer(envVars.KafkaBrokers, envVars.KafkaConsumerGroupId, envVars.KafkaIncomingTopic)
	if err != nil {
		setupLog.Error(err, "creating redpanda consumer")
		panic(err)
	}
	consumer.SetupLogger(logging.NewZapLogger(types.NamespacedName{}))
	defer consumer.Close()

	go agent.Run(consumer, myLogger)

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		panic(err)
	}
}
