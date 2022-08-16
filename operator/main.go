package main

import (
	"flag"
	"os"

	artifactsControllers "operators.kloudlite.io/controllers/artifacts"
	elasticsearchControllers "operators.kloudlite.io/controllers/elasticsearch.msvc"
	influxDbControllers "operators.kloudlite.io/controllers/influxdb.msvc"
	mongodbStandaloneControllers "operators.kloudlite.io/controllers/mongodb-standalone.msvc"
	mysqlStandaloneControllers "operators.kloudlite.io/controllers/mysql-standalone.msvc"
	opensearchControllers "operators.kloudlite.io/controllers/opensearch.msvc"
	redisStandaloneControllers "operators.kloudlite.io/controllers/redis-standalone.msvc"
	s3awsControllers "operators.kloudlite.io/controllers/s3.aws"
	serverlessControllers "operators.kloudlite.io/controllers/serverless"
	watchercontrollers "operators.kloudlite.io/controllers/watcher"
	"operators.kloudlite.io/env"
	rApi "operators.kloudlite.io/lib/operator"

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

	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	elasticsearchmsvcv1 "operators.kloudlite.io/apis/elasticsearch.msvc/v1"
	influxdbmsvcv1 "operators.kloudlite.io/apis/influxdb.msvc/v1"
	mongodbCluster "operators.kloudlite.io/apis/mongodb-cluster.msvc/v1"
	mongodbStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	mongodbexternalv1 "operators.kloudlite.io/apis/mongodb.external/v1"
	mysqlclustermsvcv1 "operators.kloudlite.io/apis/mysql-cluster.msvc/v1"
	mysqlstandalonemsvcv1 "operators.kloudlite.io/apis/mysql-standalone.msvc/v1"
	mysqlexternalv1 "operators.kloudlite.io/apis/mysql.external/v1"
	opensearchmsvcv1 "operators.kloudlite.io/apis/opensearch.msvc/v1"
	redisclustermsvcv1 "operators.kloudlite.io/apis/redis-cluster.msvc/v1"
	redisstandalonemsvcv1 "operators.kloudlite.io/apis/redis-standalone.msvc/v1"
	redpandamsvcv1 "operators.kloudlite.io/apis/redpanda.msvc/v1"
	s3awsv1 "operators.kloudlite.io/apis/s3.aws/v1"
	serverlessv1 "operators.kloudlite.io/apis/serverless/v1"
	mongodbexternalcontrollers "operators.kloudlite.io/controllers/mongodb.external"
	mysqlexternalcontrollers "operators.kloudlite.io/controllers/mysql.external"
	redpandamsvccontrollers "operators.kloudlite.io/controllers/redpanda.msvc"
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
	utilruntime.Must(redpandamsvcv1.AddToScheme(scheme))
	utilruntime.Must(mongodbexternalv1.AddToScheme(scheme))
	utilruntime.Must(mysqlexternalv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "<nothing>"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	// time.AfterFunc(
	// 	time.Second*10, func() {
	// 		// MemProfileAllocs changes which type of memory to profile
	// 		// allocations.
	// 		defer profile.Start(profile.MemProfile).Stop()
	// 	},
	// )

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var isDev bool
	var devServerHost string
	var enableForArgs arrayFlags
	var skipControllerArgs arrayFlags
	var isAllEnabled bool

	// flag.StringVar(&metricsAddr, "metrics-bind-address", ":9091", "The address the metric endpoint binds to.")
	// flag.StringVar(&probeAddr, "health-probe-bind-address", ":9092", "The address the probe endpoint binds to.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":12345", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":12346", "The address the probe endpoint binds to.")
	flag.BoolVar(
		&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.",
	)

	flag.BoolVar(&isDev, "dev", false, "--dev")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)

	flag.StringVar(&devServerHost, "serverHost", "localhost:8080", "--serverHost <host:port>")
	flag.Var(&enableForArgs, "for", "--for item1 --for item2")
	flag.Var(&skipControllerArgs, "skip", "--skip item1 --skip item2")
	flag.BoolVar(&isAllEnabled, "all", false, "--for")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	logger := logging.NewOrDie(&logging.Options{Dev: isDev})

	mgr, err := func() (manager.Manager, error) {
		cOpts := ctrl.Options{
			Scheme:                     scheme,
			MetricsBindAddress:         metricsAddr,
			Port:                       9443,
			HealthProbeBindAddress:     probeAddr,
			LeaderElection:             enableLeaderElection,
			LeaderElectionID:           "operator.kloudlite.io",
			LeaderElectionResourceLock: "configmaps",
		}
		if isDev {
			// cOpts.LeaderElectionID = "nxtcoder17.dev.kloudlite.io"
			return ctrl.NewManager(&rest.Config{Host: devServerHost}, cOpts)
		}
		return ctrl.NewManager(ctrl.GetConfigOrDie(), cOpts)
	}()
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	envVars := env.GetEnvOrDie()

	controllers := []rApi.Reconciler{
		&crds.ProjectReconciler{Name: "project"},
		&crds.AppReconciler{Name: "app"},
		&crds.RouterReconciler{Name: "router"},
		&crds.ManagedServiceReconciler{Name: "msvc"},
		&crds.ManagedResourceReconciler{Name: "mres"},

		&mongodbStandaloneControllers.ServiceReconciler{Name: "msvc-mongodb-service"},
		&mongodbStandaloneControllers.DatabaseReconciler{Name: "msvc-mongodb-database"},

		&mysqlStandaloneControllers.ServiceReconciler{Name: "msvc-mysql-service"},
		&mysqlStandaloneControllers.DatabaseReconciler{Name: "msvc-mysql-database"},

		&redisStandaloneControllers.ServiceReconciler{Name: "msvc-redis-service"},
		&redisStandaloneControllers.ACLAccountReconciler{Name: "msvc-redis-aclaccount"},

		&redpandamsvccontrollers.ServiceReconciler{Name: "msvc-redpanda-service"},
		&redpandamsvccontrollers.TopicReconciler{Name: "msvc-redpanda-topic"},

		&serverlessControllers.LambdaReconciler{Name: "lambda"},

		&elasticsearchControllers.ServiceReconciler{Name: "msvc-elasticsearch-service"},
		&opensearchControllers.ServiceReconciler{Name: "msvc-opensearch-service"},

		&influxDbControllers.ServiceReconciler{Name: "msvc-influxdb-service"},
		&influxDbControllers.BucketReconciler{Name: "msvc-influxdb-bucket"},

		&s3awsControllers.BucketReconciler{Name: "s3-aws-bucket"},

		&artifactsControllers.HarborProjectReconciler{Name: "artifacts-harbor-project"},
		&artifactsControllers.HarborUserAccountReconciler{Name: "artifacts-harbor-user-account"},

		&mongodbexternalcontrollers.DatabaseReconciler{Name: "external-mongodb-database"},
		&mysqlexternalcontrollers.DatabaseReconciler{Name: "external-mysql-database"},
	}

	producer, err := redpanda.NewProducer(envVars.KafkaBrokers)
	if err != nil {
		setupLog.Error(err, "creating redpanda producer")
		panic(err)
	}
	defer producer.Close()

	statusNotifier := watchercontrollers.NewNotifier(envVars.ClusterId, producer, envVars.KafkaStatusReplyTopic)
	billingNotifier := watchercontrollers.NewNotifier(envVars.ClusterId, producer, envVars.KafkaBillingReplyTopic)

	controllers = append(
		controllers,
		&watchercontrollers.StatusWatcherReconciler{Name: "status-watcher", Notifier: statusNotifier},
		&watchercontrollers.BillingWatcherReconciler{Name: "billing-watcher", Notifier: billingNotifier},
	)

	enabledForControllers := map[string]bool{}
	for _, arg := range enableForArgs {
		enabledForControllers[arg] = true
	}

	skippedControllers := map[string]bool{}
	for _, arg := range skipControllerArgs {
		skippedControllers[arg] = true
	}

	for _, rc := range controllers {
		if skippedControllers[rc.GetName()] {
			logger.Infof("skipping %s controller (by flag) ", rc.GetName())
			continue
		}
		if isAllEnabled || enabledForControllers[rc.GetName()] {
			if err := rc.SetupWithManager(mgr, envVars, logger); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", rc.GetName())
				os.Exit(1)
			}
		}
	}

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		panic(err)
	}
}
