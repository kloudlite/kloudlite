package operator

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/kloudlite/operator/toolkit/kubectl"
	"github.com/kloudlite/operator/toolkit/logging"
	reconciler "github.com/kloudlite/operator/toolkit/reconciler"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
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
}

type Operator interface {
	AddToSchemes(fns ...func(s *runtime.Scheme) error)
	RegisterControllers(controllers ...reconciler.Reconciler)
	// RegisterWebhooks(objects ...WebhookEnabledType)
	Start()
	Operator() *operator
}

type operator struct {
	startedAt  time.Time
	mgrConfig  *rest.Config
	mgrOptions ctrl.Options

	controllers []func(mgr manager.Manager)
	webhooks    []func(mgr manager.Manager)

	registeredControllers map[string]struct{}

	IsDev        bool
	schemesAdded bool
	Scheme       *runtime.Scheme

	k8sYamlClient kubectl.YAMLClient
	logger        *slog.Logger
}

func (op *operator) KubeYAMLClient() kubectl.YAMLClient {
	return op.k8sYamlClient
}

func (op *operator) Logger() *slog.Logger {
	return op.logger
}

func New(name string) Operator {
	printBuildInfo()

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var isDev bool
	var debugLog bool
	var devServerHost string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":12345", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":12346", "The address the probe endpoint binds to.")
	flag.BoolVar(
		&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controllers manager. "+
			"Enabling this will ensure there is only one active controllers manager.",
	)

	opts := zap.Options{
		Development: true,
		EncoderConfigOptions: []zap.EncoderConfigOption{
			func(ec *zapcore.EncoderConfig) {
				// ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
				ec.TimeKey = ""
			},
		},
	}

	opts.BindFlags(flag.CommandLine)
	rest.SetDefaultWarningHandler(rest.NoWarnings{})

	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.BoolVar(&debugLog, "debug-log", false, "--debug-log")
	flag.StringVar(&devServerHost, "serverHost", "localhost:8080", "--serverHost <host:port>")
	flag.Parse()

	if debugLog {
		os.Setenv("LOG_DEBUG", "true")
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgrConfig, mgrOptions := func() (*rest.Config, ctrl.Options) {
		cOpts := ctrl.Options{
			Scheme:                        scheme,
			LeaderElection:                enableLeaderElection,
			LeaderElectionReleaseOnCancel: true,
			LeaderElectionID:              fmt.Sprintf("operator-%s.kloudlite.io", name),
			LeaderElectionResourceLock:    "leases",
		}
		if isDev {
			cOpts.Metrics.BindAddress = "0"
			ctrl.Log.Info("dev mode enabled, using", "server-host", devServerHost)
			return &rest.Config{Host: devServerHost}, cOpts
		}

		cOpts.Metrics.BindAddress = metricsAddr
		cOpts.HealthProbeBindAddress = probeAddr
		return ctrl.GetConfigOrDie(), cOpts
	}()

	logger := logging.New(ctrl.Log)
	k8sYamlClient, err := kubectl.NewYAMLClient(mgrConfig, kubectl.YAMLClientOpts{Logger: logger})
	if err != nil {
		log.Fatalln(err)
	}

	return &operator{
		startedAt:     time.Now(),
		mgrConfig:     mgrConfig,
		mgrOptions:    mgrOptions,
		IsDev:         isDev,
		k8sYamlClient: k8sYamlClient,
		logger:        logger,
	}
}

func (op *operator) AddToSchemes(fns ...func(s *runtime.Scheme) error) {
	for i := range fns {
		utilruntime.Must(fns[i](scheme))
	}

	op.mgrOptions.Scheme = scheme
}

func (op *operator) RegisterControllers(controllers ...reconciler.Reconciler) {
	for i := range controllers {
		controller := controllers[i]
		op.controllers = append(op.controllers, func(mgr manager.Manager) {
			_, ok := op.registeredControllers[controller.GetName()]
			if ok {
				return
				// ctrl.Log.Info("controller %s already registered, skipping", controller.GetName())
			}
			if op.registeredControllers == nil {
				op.registeredControllers = make(map[string]struct{})
			}
			op.registeredControllers[controller.GetName()] = struct{}{}
			setupLog.Info("registering controller", "controller", controller.GetName())
			// go func() {
			// controller.Client = mgr.GetClient()
			if err := controller.SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controllers", "controllers", controller.GetName())
				os.Exit(1)
			}
			// }()
		})
	}
}

// type WebhookEnabledType interface {
// 	client.Object
// 	SetupWebhookWithManager(mgr ctrl.Manager) error
// }
//
// func (op *operator) RegisterWebhooks(types ...WebhookEnabledType) {
// 	for i := range types {
// 		webhookType := types[i]
// 		op.webhooks = append(op.webhooks, func(mgr manager.Manager) {
// 			setupLog.Info("registering webhook", "for", webhookType.GetName())
// 			if err := webhookType.SetupWebhookWithManager(mgr); err != nil {
// 				setupLog.Error(err, "unable to create webhook", "webhook", types[i].GetName())
// 				os.Exit(1)
// 			}
// 		})
// 	}
// }

func (op *operator) Operator() *operator {
	return op
}

func (op *operator) Start() {
	mgr, err := ctrl.NewManager(op.mgrConfig, op.mgrOptions)
	if err != nil {
		log.Fatalln(err)
	}

	for i := range op.controllers {
		op.controllers[i](mgr)
	}

	// for i := range op.webhooks {
	// 	op.webhooks[i](mgr)
	// }

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	printReadyBanner(time.Since(op.startedAt))
	ctrl.Log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		panic(err)
	}
}
