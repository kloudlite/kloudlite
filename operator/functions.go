package operator

import (
	"flag"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	csiv1 "operators.kloudlite.io/apis/csi/v1"
	elasticsearchmsvcv1 "operators.kloudlite.io/apis/elasticsearch.msvc/v1"
	extensionsv1 "operators.kloudlite.io/apis/extensions/v1"
	influxdbmsvcv1 "operators.kloudlite.io/apis/influxdb.msvc/v1"
	mongodbCluster "operators.kloudlite.io/apis/mongodb-cluster.msvc/v1"
	mongodbStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	mongodbexternalv1 "operators.kloudlite.io/apis/mongodb.external/v1"
	mongodbMsvcv1 "operators.kloudlite.io/apis/mongodb.msvc/v1"
	mysqlclustermsvcv1 "operators.kloudlite.io/apis/mysql-cluster.msvc/v1"
	mysqlstandalonemsvcv1 "operators.kloudlite.io/apis/mysql-standalone.msvc/v1"
	mysqlexternalv1 "operators.kloudlite.io/apis/mysql.external/v1"
	mysqlMsvcv1 "operators.kloudlite.io/apis/mysql.msvc/v1"
	neo4jMsvcv1 "operators.kloudlite.io/apis/neo4j.msvc/v1"
	opensearchmsvcv1 "operators.kloudlite.io/apis/opensearch.msvc/v1"
	redisclustermsvcv1 "operators.kloudlite.io/apis/redis-cluster.msvc/v1"
	redisstandalonemsvcv1 "operators.kloudlite.io/apis/redis-standalone.msvc/v1"
	redisMsvcv1 "operators.kloudlite.io/apis/redis.msvc/v1"
	redpandamsvcv1 "operators.kloudlite.io/apis/redpanda.msvc/v1"
	s3awsv1 "operators.kloudlite.io/apis/s3.aws/v1"

	// serverlessv1 "operators.kloudlite.io/apis/serverless/v1"
	zookeeperMsvcv1 "operators.kloudlite.io/apis/zookeeper.msvc/v1"
	flagTypes "operators.kloudlite.io/lib/flag-types"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("operator")
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
	// utilruntime.Must(serverlessv1.AddToScheme(scheme))
	utilruntime.Must(elasticsearchmsvcv1.AddToScheme(scheme))
	utilruntime.Must(opensearchmsvcv1.AddToScheme(scheme))
	utilruntime.Must(s3awsv1.AddToScheme(scheme))
	utilruntime.Must(artifactsv1.AddToScheme(scheme))
	utilruntime.Must(redpandamsvcv1.AddToScheme(scheme))
	utilruntime.Must(mongodbexternalv1.AddToScheme(scheme))
	utilruntime.Must(mysqlexternalv1.AddToScheme(scheme))
	utilruntime.Must(mongodbMsvcv1.AddToScheme(scheme))
	utilruntime.Must(mysqlMsvcv1.AddToScheme(scheme))
	utilruntime.Must(redisMsvcv1.AddToScheme(scheme))
	utilruntime.Must(zookeeperMsvcv1.AddToScheme(scheme))
	utilruntime.Must(extensionsv1.AddToScheme(scheme))
	utilruntime.Must(neo4jMsvcv1.AddToScheme(scheme))
	utilruntime.Must(csiv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

type operator struct {
	Manager            manager.Manager
	Logger             logging.Logger
	enableForArgs      flagTypes.StringArray
	skipControllerArgs flagTypes.StringArray
	isAllEnabled       bool
	IsDev              bool
}

func New(name string) *operator {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var isDev bool
	var devServerHost string
	var enableForArgs flagTypes.StringArray
	var skipControllerArgs flagTypes.StringArray
	// var isAllEnabled bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":12345", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":12346", "The address the probe endpoint binds to.")
	flag.BoolVar(
		&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controllers manager. "+
			"Enabling this will ensure there is only one active controllers manager.",
	)

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)

	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.StringVar(&devServerHost, "serverHost", "localhost:8080", "--serverHost <host:port>")
	flag.Var(&enableForArgs, "for", "--for item1 --for item2")
	flag.Var(&skipControllerArgs, "skip", "--skip item1 --skip item2")
	// flag.BoolVar(&isAllEnabled, "all", true, "--all")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	logger := logging.NewOrDie(&logging.Options{Dev: isDev})

	mgr, err := func() (manager.Manager, error) {
		cOpts := ctrl.Options{
			Scheme:                     scheme,
			Port:                       9443,
			LeaderElection:             enableLeaderElection,
			LeaderElectionID:           fmt.Sprintf("operator-%s.kloudlite.io", name),
			LeaderElectionResourceLock: "configmaps",
		}

		if isDev {
			cOpts.MetricsBindAddress = "0"
			// cOpts.LeaderElectionID = "nxtcoder17.dev.kloudlite.io"
			return ctrl.NewManager(&rest.Config{Host: devServerHost}, cOpts)
		}

		cOpts.MetricsBindAddress = metricsAddr
		cOpts.HealthProbeBindAddress = probeAddr
		return ctrl.NewManager(ctrl.GetConfigOrDie(), cOpts)
	}()
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	return &operator{
		Manager:            mgr,
		Logger:             logger,
		enableForArgs:      enableForArgs,
		skipControllerArgs: skipControllerArgs,
		IsDev:              isDev,
	}
}

func (op *operator) RegisterControllers(controllers ...rApi.Reconciler) {
	// enabledForControllers := map[string]bool{}
	// for _, arg := range op.enableForArgs {
	// 	enabledForControllers[arg] = true
	// }
	//
	// skippedControllers := map[string]bool{}
	// for _, arg := range op.skipControllerArgs {
	// 	skippedControllers[arg] = true
	// }

	for _, rc := range controllers {
		// if skippedControllers[rc.GetName()] {
		// 	setupLog.Info(fmt.Sprintf("skipping %s controllers (by flag) ", rc.GetName()))
		// 	continue
		// }
		//
		// if enabledForControllers[rc.GetName()] {
		// 	if err := rc.SetupWithManager(op.Manager, op.Env, op.Logger); err != nil {
		// 		setupLog.Error(err, "unable to create controllers", "controllers", rc.GetName())
		// 		os.Exit(1)
		// 	}
		// }
		if err := rc.SetupWithManager(op.Manager, op.Logger); err != nil {
			setupLog.Error(err, "unable to create controllers", "controllers", rc.GetName())
			os.Exit(1)
		}
	}

	if err := op.Manager.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := op.Manager.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
}

func (op *operator) Start() {
	fmt.Println(
		`
██████  ███████  █████  ██████  ██    ██ 
██   ██ ██      ██   ██ ██   ██  ██  ██  
██████  █████   ███████ ██   ██   ████   
██   ██ ██      ██   ██ ██   ██    ██    
██   ██ ███████ ██   ██ ██████     ██    
	`,
	)

	op.Logger.Infof("starting manager")
	if err := op.Manager.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		panic(err)
	}
}
