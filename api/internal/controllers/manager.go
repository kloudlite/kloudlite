package controllers

import (
	"context"
	"fmt"

	"github.com/kloudlite/kloudlite/api/internal/controllers/helmchart"
	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	interceptsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/intercepts/v1"
	machinesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/machines/v1"
	packagesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/packages/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/pkg/apis/platform/v1alpha1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	zaplog "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

type Manager struct {
	mgr    ctrl.Manager
	logger *zap.Logger
}

// NewManager creates a new controller manager with all controllers
func NewManager(cfg *rest.Config, logger *zap.Logger) (*Manager, error) {
	// Setup scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(platformv1alpha1.AddToScheme(scheme))
	utilruntime.Must(machinesv1.AddToScheme(scheme))
	utilruntime.Must(environmentsv1.AddToScheme(scheme))
	utilruntime.Must(workspacesv1.AddToScheme(scheme))
	utilruntime.Must(packagesv1.AddToScheme(scheme))
	utilruntime.Must(interceptsv1.AddToScheme(scheme))
	utilruntime.Must(metricsv1beta1.AddToScheme(scheme))

	// Set controller-runtime logger
	ctrl.SetLogger(zaplog.New(func(o *zaplog.Options) {
		o.Development = true
	}))

	// Create manager
	// Disable metrics to avoid port conflict with main server
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: "", // Disable health probe
		Metrics: server.Options{
			BindAddress: "0", // Disable metrics server
		},
		LeaderElection:   false,
		LeaderElectionID: "kloudlite-api-controller-manager",
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create manager: %w", err)
	}

	// Health checks disabled to avoid port conflicts

	// Setup User controller
	userReconciler := &UserReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: logger.With(zap.String("controller", "user")),
	}

	if err = userReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create User controller: %w", err)
	}

	// Setup Environment controller
	environmentReconciler := &EnvironmentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: logger.With(zap.String("controller", "environment")),
	}

	if err = environmentReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create Environment controller: %w", err)
	}

	// Setup WorkMachine controller
	workMachineReconciler := &WorkMachineReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	if err = workMachineReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create WorkMachine controller: %w", err)
	}

	// Setup Composition controller
	compositionReconciler := &CompositionReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: logger.With(zap.String("controller", "composition")),
	}

	if err = compositionReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create Composition controller: %w", err)
	}

	// Setup Workspace controller
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create kubernetes clientset: %w", err)
	}

	workspaceReconciler := &WorkspaceReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Logger:    logger.With(zap.String("controller", "workspace")),
		Config:    cfg,
		Clientset: clientset,
	}

	if err = workspaceReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create Workspace controller: %w", err)
	}

	// Setup HelmChart controller
	helmChartReconciler := &helmchart.Reconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		HelmJobRunnerImage: "ghcr.io/kloudlite/plugin-helm-chart/helm-job-runner:v1.0.0",
	}

	if err = helmChartReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create HelmChart controller: %w", err)
	}

	// Setup ServiceIntercept controller
	serviceInterceptReconciler := &ServiceInterceptReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: logger.With(zap.String("controller", "serviceintercept")),
	}

	if err = serviceInterceptReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create ServiceIntercept controller: %w", err)
	}

	logger.Info("Controllers initialized successfully")

	return &Manager{
		mgr:    mgr,
		logger: logger,
	}, nil
}

// Start starts the controller manager
func (m *Manager) Start(ctx context.Context) error {
	m.logger.Info("Starting controller manager")
	return m.mgr.Start(ctx)
}

// GetClient returns the controller-runtime client
func (m *Manager) GetClient() client.Client {
	return m.mgr.GetClient()
}
