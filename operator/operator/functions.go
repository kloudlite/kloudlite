package operator

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kloudlite/operator/common"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	RegisterControllers(controllers ...rApi.Reconciler)
	RegisterWebhooks(objects ...WebhookEnabledType)
	Start()
	Operator() *operator
}

type operator struct {
	mgrConfig  *rest.Config
	mgrOptions ctrl.Options

	controllers []func(mgr manager.Manager)
	webhooks    []func(mgr manager.Manager)

	registeredControllers map[string]struct{}

	Logger        logging.Logger
	IsDev         bool
	schemesAdded  bool
	Scheme        *runtime.Scheme
	k8sYamlClient kubectl.YAMLClient
}

func New(name string) Operator {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var isDev bool
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
	}
	opts.BindFlags(flag.CommandLine)
	rest.SetDefaultWarningHandler(rest.NoWarnings{})

	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.StringVar(&devServerHost, "serverHost", "localhost:8080", "--serverHost <host:port>")
	// flag.BoolVar(&isAllEnabled, "all", true, "--all")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	logger := logging.NewOrDie(&logging.Options{Dev: true})

	mgrConfig, mgrOptions := func() (*rest.Config, ctrl.Options) {
		cOpts := ctrl.Options{
			Scheme:                     scheme,
			Port:                       9443,
			LeaderElection:             enableLeaderElection,
			LeaderElectionID:           fmt.Sprintf("operator-%s.kloudlite.io", name),
			LeaderElectionResourceLock: "configmapsleases",
		}
		if isDev {
			cOpts.MetricsBindAddress = "0"
			// cOpts.MetricsBindAddress = "0"
			// cOpts.LeaderElectionID = "nxtcoder17.dev.kloudlite.io"
			logger.Warnf("dev mode enabled, using dev server host: %s", devServerHost)
			return &rest.Config{Host: devServerHost}, cOpts
		}

		cOpts.MetricsBindAddress = metricsAddr
		cOpts.HealthProbeBindAddress = probeAddr
		return ctrl.GetConfigOrDie(), cOpts
	}()

	k8sYamlClient, err := kubectl.NewYAMLClient(mgrConfig)
	if err != nil {
		log.Fatalln(err)
	}

	return &operator{
		mgrConfig:     mgrConfig,
		mgrOptions:    mgrOptions,
		Logger:        logger,
		IsDev:         isDev,
		k8sYamlClient: k8sYamlClient,
	}
}

func (op *operator) AddToSchemes(fns ...func(s *runtime.Scheme) error) {
	for i := range fns {
		utilruntime.Must(fns[i](scheme))
	}

	op.mgrOptions.Scheme = scheme
}

func (op *operator) RegisterControllers(controllers ...rApi.Reconciler) {
	for i := range controllers {
		controller := controllers[i]
		op.controllers = append(op.controllers, func(mgr manager.Manager) {
			_, ok := op.registeredControllers[controller.GetName()]
			if ok {
				op.Logger.Debugf("controller %s already registered, skipping", controller.GetName())
			}
			if !ok {
				if op.registeredControllers == nil {
					op.registeredControllers = make(map[string]struct{})
				}
				op.registeredControllers[controller.GetName()] = struct{}{}
				setupLog.Info("registering controller", "controller", controller.GetName())
				go func() {
					if err := controller.SetupWithManager(mgr, op.Logger); err != nil {
						setupLog.Error(err, "unable to create controllers", "controllers", controller.GetName())
						os.Exit(1)
					}
				}()
			}
		})
	}
}

type WebhookEnabledType interface {
	client.Object
	SetupWebhookWithManager(mgr ctrl.Manager) error
}

func (op *operator) RegisterWebhooks(types ...WebhookEnabledType) {
	for i := range types {
		webhookType := types[i]
		op.webhooks = append(op.webhooks, func(mgr manager.Manager) {
			setupLog.Info("registering webhook", "for", webhookType.GetName())
			if err := webhookType.SetupWebhookWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create webhook", "webhook", types[i].GetName())
				os.Exit(1)
			}
		})
	}
}

func (op *operator) Operator() *operator {
	return op
}

func (op *operator) Start() {
	op.Logger.Infof("starting manager")

	mgr, err := ctrl.NewManager(op.mgrConfig, op.mgrOptions)
	if err != nil {
		log.Fatalln(err)
	}

	for i := range op.controllers {
		op.controllers[i](mgr)
	}

	for i := range op.webhooks {
		op.webhooks[i](mgr)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	common.PrintReadyBanner()
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		panic(err)
	}
}
