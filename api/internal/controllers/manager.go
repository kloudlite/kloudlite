package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/zapr"
	"github.com/kloudlite/kloudlite/api/internal/config"
	ca "github.com/kloudlite/kloudlite/api/internal/controllers/certs"
	"github.com/kloudlite/kloudlite/api/internal/controllers/composition"
	connectiontokenv1 "github.com/kloudlite/kloudlite/api/internal/controllers/connectiontoken/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest"
	domainrequestsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/environment"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/user"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workspace"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
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
	mgr    ctrl.Manager
	logger *zap.Logger
}

// NewManager creates a new controller manager with all controllers
func NewManager(cfg *rest.Config, installationCfg *config.InstallationConfig, logger *zap.Logger) (*Manager, error) {
	// Setup scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(platformv1alpha1.AddToScheme(scheme))
	utilruntime.Must(machinesv1.AddToScheme(scheme))
	utilruntime.Must(environmentsv1.AddToScheme(scheme))
	utilruntime.Must(workspacev1.AddToScheme(scheme))
	utilruntime.Must(packagesv1.AddToScheme(scheme))
	utilruntime.Must(interceptsv1.AddToScheme(scheme))
	utilruntime.Must(connectiontokenv1.AddToScheme(scheme))
	utilruntime.Must(domainrequestsv1.AddToScheme(scheme))
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

	// Setup Environment controller
	environmentReconciler := &environment.EnvironmentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: logger.With(zap.String("controller", "environment")),
	}

	if err = environmentReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create Environment controller: %w", err)
	}

	if err := workmachine.Register(mgr); err != nil {
		return nil, fmt.Errorf("unable to setup WorkMachine controller: %w", err)
	}

	if err := ca.Register(mgr); err != nil {
		return nil, fmt.Errorf("unable to setup WorkMachine controller: %w", err)
	}

	// Setup Composition controller
	compositionReconciler := &composition.CompositionReconciler{
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

	workspaceReconciler := &workspace.WorkspaceReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Logger:    logger.With(zap.String("controller", "workspace")),
		Config:    cfg,
		Clientset: clientset,
	}

	if err = workspaceReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create Workspace controller: %w", err)
	}

	// Setup ServiceIntercept controller
	serviceInterceptReconciler := &serviceintercept.ServiceInterceptReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: logger.With(zap.String("controller", "serviceintercept")),
	}

	if err = serviceInterceptReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create ServiceIntercept controller: %w", err)
	}

	// Setup DomainRequest controller
	domainRequestReconciler := &domainrequest.DomainRequestReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		Logger:             logger.With(zap.String("controller", "domainrequest")),
		InstallationKey:    installationCfg.InstallationKey,
		InstallationSecret: installationCfg.InstallationSecret,
	}

	if err = domainRequestReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to create DomainRequest controller: %w", err)
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
