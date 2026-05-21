package controllers

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/zapr"
	"github.com/kloudlite/kloudlite/api/config"
	"github.com/kloudlite/kloudlite/controllers/environment"
	"github.com/kloudlite/kloudlite/controllers/snapshot"
	"github.com/kloudlite/kloudlite/controllers/user"
	"github.com/kloudlite/kloudlite/controllers/wmingress"
	"github.com/kloudlite/kloudlite/controllers/workmachine/machinescoped"
	"github.com/kloudlite/kloudlite/controllers/workmachine/platformscoped"
	"github.com/kloudlite/kloudlite/controllers/workspace"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	packagesv1 "github.com/kloudlite/kloudlite/types/packages/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/types/snapshot/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/types/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/types/workspace/v1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

type Manager struct {
	mgr               ctrl.Manager
	logger            *zap.Logger
	ingressReconciler interface {
		StartServers(context.Context) error
	}
}

func PlatformAPIControllerNames() []string {
	return []string{"user", "workmachine-platform-scoped"}
}

func MachineScopedControllerNames() []string {
	return []string{"workmachine-manager", "environment", "workspace"}
}

// NewManager creates a new controller manager with all controllers
func NewManager(cfg *rest.Config, installationCfg *config.InstallationConfig, authCfg *config.AuthConfig, logger *zap.Logger) (*Manager, error) {
	return newManager(cfg, installationCfg, authCfg, logger, false)
}

func NewPlatformAPIManager(cfg *rest.Config, installationCfg *config.InstallationConfig, authCfg *config.AuthConfig, logger *zap.Logger) (*Manager, error) {
	return newManager(cfg, installationCfg, authCfg, logger, true)
}

func NewMachineScopedManager(cfg *rest.Config, installationCfg *config.InstallationConfig, authCfg *config.AuthConfig, logger *zap.Logger) (*Manager, error) {
	return newMachineScopedManager(cfg, installationCfg, authCfg, logger)
}

func newMachineScopedManager(cfg *rest.Config, installationCfg *config.InstallationConfig, authCfg *config.AuthConfig, logger *zap.Logger) (*Manager, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(platformv1alpha1.AddToScheme(scheme))
	utilruntime.Must(machinesv1.AddToScheme(scheme))
	utilruntime.Must(environmentsv1.AddToScheme(scheme))
	utilruntime.Must(workspacev1.AddToScheme(scheme))
	utilruntime.Must(packagesv1.AddToScheme(scheme))
	utilruntime.Must(snapshotv1.AddToScheme(scheme))
	utilruntime.Must(metricsv1beta1.AddToScheme(scheme))

	ctrl.SetLogger(zapr.NewLogger(logger))

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: "",
		Metrics:                server.Options{BindAddress: "0"},
		LeaderElection:         false,
		LeaderElectionID:       "kloudlite-workmachine-manager",
		WebhookServer:          nil,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create workmachine-manager: %w", err)
	}

	controllerCfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to load controller configuration: %w", err)
	}
	ingressReconciler := &wmingress.IngressReconciler{
		Client:                  mgr.GetClient(),
		Scheme:                  mgr.GetScheme(),
		Logger:                  logger.With(zap.String("controller", "wm-ingress")),
		HTTPPort:                controllerCfg.WMIngress.HTTPPort,
		HTTPSPort:               controllerCfg.WMIngress.HTTPSPort,
		WildcardDomain:          controllerCfg.WMIngress.WildcardDomain,
		WildcardSecretName:      controllerCfg.WMIngress.WildcardSecretName,
		WildcardSecretNamespace: controllerCfg.WMIngress.WildcardSecretNamespace,
		OwnNamespace:            os.Getenv("POD_NAMESPACE"),
		RegistryUsername:        controllerCfg.WMIngress.RegistryUsername,
		ForceFullRebuild:        controllerCfg.WMIngress.ForceFullRebuild,
	}
	if ingressReconciler.HTTPPort == 0 {
		ingressReconciler.HTTPPort = 80
	}
	if ingressReconciler.HTTPSPort == 0 {
		ingressReconciler.HTTPSPort = 443
	}
	if ingressReconciler.WildcardSecretName == "" {
		ingressReconciler.WildcardSecretName = "kloudlite-wildcard-cert-tls"
	}
	if ingressReconciler.WildcardSecretNamespace == "" {
		ingressReconciler.WildcardSecretNamespace = os.Getenv("WM_INGRESS_WILDCARD_SECRET_NAMESPACE")
	}
	if ingressReconciler.WildcardSecretNamespace == "" {
		ingressReconciler.WildcardSecretNamespace = os.Getenv("POD_NAMESPACE")
	}
	if ingressReconciler.OwnNamespace == "" {
		ingressReconciler.OwnNamespace = os.Getenv("WM_INGRESS_OWN_NAMESPACE")
	}
	if ingressReconciler.OwnNamespace == "" {
		ingressReconciler.OwnNamespace = os.Getenv("POD_NAMESPACE")
	}
	if ingressReconciler.RegistryUsername == "" {
		ingressReconciler.RegistryUsername = os.Getenv("REGISTRY_USERNAME")
	}
	if err := ingressReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to setup integrated wm-ingress controller: %w", err)
	}

	environmentReconciler := &environment.EnvironmentReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		Logger:          logger.With(zap.String("controller", "environment")),
		Cfg:             controllerCfg,
		OwnNamespace:    os.Getenv("POD_NAMESPACE"),
		WorkMachineName: os.Getenv("WORKMACHINE_NAME"),
	}
	if environmentReconciler.OwnNamespace == "" {
		environmentReconciler.OwnNamespace = os.Getenv("NAMESPACE")
	}
	if err := environmentReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to setup scoped Environment controller: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create kubernetes clientset: %w", err)
	}

	workspaceCfg := &workspace.ControllerConfig{}
	workspaceCfg.Workspace.DefaultIdleTimeoutMinutes = controllerCfg.Workspace.DefaultIdleTimeoutMinutes
	workspaceCfg.Workspace.RequeueIntervalMinutes = controllerCfg.Workspace.RequeueIntervalMinutes
	workspaceCfg.Workspace.RBACCleanupIntervalMinutes = controllerCfg.Workspace.RBACCleanupIntervalMinutes
	workspaceCfg.Workspace.KubectlImage = controllerCfg.Workspace.KubectlImage
	workspaceCfg.Workspace.GitImage = controllerCfg.Workspace.GitImage
	workspaceCfg.Workspace.AlpineImage = controllerCfg.Workspace.AlpineImage
	workspaceCfg.Workspace.CleanupPodTTLSeconds = controllerCfg.Workspace.CleanupPodTTLSeconds
	workspaceCfg.Workspace.VSCodeVersion = controllerCfg.Workspace.VSCodeVersion
	workspaceCfg.Environment.LifecycleRetryInterval = controllerCfg.Environment.LifecycleRetryInterval

	workspaceReconciler := &workspace.WorkspaceReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		Logger:          logger.With(zap.String("controller", "workspace")),
		Config:          cfg,
		Clientset:       clientset,
		JWTSecret:       authCfg.JWTSecret,
		Cfg:             workspaceCfg,
		OwnNamespace:    os.Getenv("POD_NAMESPACE"),
		WorkMachineName: os.Getenv("WORKMACHINE_NAME"),
	}
	if workspaceReconciler.OwnNamespace == "" {
		workspaceReconciler.OwnNamespace = os.Getenv("NAMESPACE")
	}
	if err := workspaceReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to setup scoped Workspace controller: %w", err)
	}

	if err := machinescoped.Register(mgr, controllerCfg); err != nil {
		return nil, fmt.Errorf("unable to setup WorkMachine manager controller: %w", err)
	}

	logger.Info("controllers initialized successfully", zap.Strings("controllers", MachineScopedControllerNames()))
	return &Manager{mgr: mgr, logger: logger, ingressReconciler: ingressReconciler}, nil
}

func newManager(cfg *rest.Config, installationCfg *config.InstallationConfig, authCfg *config.AuthConfig, logger *zap.Logger, platformAPIOnly bool) (*Manager, error) {
	// Setup scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(platformv1alpha1.AddToScheme(scheme))
	utilruntime.Must(machinesv1.AddToScheme(scheme))
	utilruntime.Must(environmentsv1.AddToScheme(scheme))
	utilruntime.Must(workspacev1.AddToScheme(scheme))
	utilruntime.Must(packagesv1.AddToScheme(scheme))
	utilruntime.Must(snapshotv1.AddToScheme(scheme))
	utilruntime.Must(metricsv1beta1.AddToScheme(scheme))

	// Set controller-runtime logger to use our zap logger
	ctrl.SetLogger(zapr.NewLogger(logger))

	// Create manager
	// Disable metrics to avoid port conflict with main server
	// Disable webhook server since webhooks are handled by main Gin server
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: "", // Disable health probe
		Metrics: server.Options{
			BindAddress: "0", // Disable metrics server
		},
		LeaderElection:   false,
		LeaderElectionID: "kloudlite-api-controller-manager",
		WebhookServer:    nil, // Disable - webhooks handled by main server
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create manager: %w", err)
	}

	// Setup field indexes for efficient queries by owner
	ctx := context.Background()

	// Index Environment by ownedBy
	if err := mgr.GetFieldIndexer().IndexField(ctx, &environmentsv1.Environment{}, "spec.ownedBy", func(obj client.Object) []string {
		env := obj.(*environmentsv1.Environment)
		return []string{env.Spec.OwnedBy}
	}); err != nil {
		return nil, fmt.Errorf("unable to create Environment ownedBy index: %w", err)
	}

	// Index Workspace by ownedBy
	if err := mgr.GetFieldIndexer().IndexField(ctx, &workspacev1.Workspace{}, "spec.ownedBy", func(obj client.Object) []string {
		ws := obj.(*workspacev1.Workspace)
		return []string{ws.Spec.OwnedBy}
	}); err != nil {
		return nil, fmt.Errorf("unable to create Workspace ownedBy index: %w", err)
	}

	// Index WorkMachine by ownedBy
	if err := mgr.GetFieldIndexer().IndexField(ctx, &machinesv1.WorkMachine{}, "spec.ownedBy", func(obj client.Object) []string {
		wm := obj.(*machinesv1.WorkMachine)
		return []string{wm.Spec.OwnedBy}
	}); err != nil {
		return nil, fmt.Errorf("unable to create WorkMachine ownedBy index: %w", err)
	}

	// Setup User controller
	userReconciler := &user.UserReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: logger.With(zap.String("controller", "user")),
	}

	if err = userReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create User controller: %w", err)
	}

	// Load controller configuration
	controllerCfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to load controller configuration: %w", err)
	}

	if platformAPIOnly {
		if err := platformscoped.Register(mgr, controllerCfg); err != nil {
			return nil, fmt.Errorf("unable to setup platform-scoped WorkMachine controller: %w", err)
		}
	} else {
		if err := platformscoped.Register(mgr, controllerCfg); err != nil {
			return nil, fmt.Errorf("unable to setup platform-scoped WorkMachine controller: %w", err)
		}
		if err := machinescoped.Register(mgr, controllerCfg); err != nil {
			return nil, fmt.Errorf("unable to setup WorkMachine manager controller: %w", err)
		}
	}

	if platformAPIOnly {
		logger.Info("controllers initialized successfully", zap.Strings("controllers", PlatformAPIControllerNames()))
		return &Manager{
			mgr:    mgr,
			logger: logger,
		}, nil
	}

	// Setup Environment controller
	environmentReconciler := &environment.EnvironmentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: logger.With(zap.String("controller", "environment")),
		Cfg:    controllerCfg,
	}

	if err = environmentReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create Environment controller: %w", err)
	}

	// Setup Snapshot controller with operator for registry operations
	snapshotOperator := snapshot.NewDefaultSnapshotOperator(logger.With(zap.String("component", "snapshot-operator")))
	snapshotReconciler := &snapshot.SnapshotReconciler{
		Client:           mgr.GetClient(),
		Logger:           logger.With(zap.String("controller", "snapshot")),
		SnapshotOperator: snapshotOperator,
	}

	if err = snapshotReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create Snapshot controller: %w", err)
	}

	// Setup SnapshotRestore controller
	snapshotRestoreReconciler := &snapshot.SnapshotRestoreReconciler{
		Client: mgr.GetClient(),
		Logger: logger.With(zap.String("controller", "snapshotrestore")),
	}

	if err = snapshotRestoreReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create SnapshotRestore controller: %w", err)
	}

	// Setup EnvironmentSnapshotRequest controller
	envSnapshotRequestReconciler := &environment.EnvironmentSnapshotRequestReconciler{
		Client: mgr.GetClient(),
		Logger: logger.With(zap.String("controller", "environmentsnapshotrequest")),
	}

	if err = envSnapshotRequestReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create EnvironmentSnapshotRequest controller: %w", err)
	}

	// Setup EnvironmentSnapshotRestore controller
	envSnapshotRestoreReconciler := &environment.EnvironmentSnapshotRestoreReconciler{
		Client: mgr.GetClient(),
		Logger: logger.With(zap.String("controller", "environmentsnapshotrestore")),
	}

	if err = envSnapshotRestoreReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create EnvironmentSnapshotRestore controller: %w", err)
	}

	// Setup EnvironmentForkRequest controller
	envForkRequestReconciler := &environment.EnvironmentForkRequestReconciler{
		Client: mgr.GetClient(),
		Logger: logger.With(zap.String("controller", "environmentforkrequest")),
	}

	if err = envForkRequestReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create EnvironmentForkRequest controller: %w", err)
	}

	logger.Info("controllers initialized successfully")

	return &Manager{
		mgr:    mgr,
		logger: logger,
	}, nil
}

// Start starts the controller manager
func (m *Manager) Start(ctx context.Context) error {
	m.logger.Info("starting controller manager")
	return m.mgr.Start(ctx)
}

func (m *Manager) StartIngressServers(ctx context.Context) error {
	if m.ingressReconciler == nil {
		return nil
	}
	return m.ingressReconciler.StartServers(ctx)
}

// GetClient returns the controller-runtime client
func (m *Manager) GetClient() client.Client {
	return m.mgr.GetClient()
}
