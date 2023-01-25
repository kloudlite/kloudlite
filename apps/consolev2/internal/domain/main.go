package domain

import (
	"kloudlite.io/apps/consolev2/internal/env"
	"text/template"

	"go.uber.org/fx"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	kldns "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/dns"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/jseval"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
	rcn "kloudlite.io/pkg/res-change-notifier"
)

type domain struct {
	dnsClient            kldns.DNSClient
	workloadMessenger    WorkloadMessenger
	providerRepo         repos.DbRepo[*entities.CloudProvider]
	financeClient        finance.FinanceClient
	consoleTemplate      *template.Template
	clusterRepo          repos.DbRepo[*entities.Cluster]
	k8sYamlClient        *k8s.YAMLClient
	secretRepo           repos.DbRepo[*entities.Secret]
	projectRepo          repos.DbRepo[*entities.Project]
	environmentRepo      repos.DbRepo[*entities.Environment]
	appRepo              repos.DbRepo[*entities.App]
	configRepo           repos.DbRepo[*entities.Config]
	routerRepo           repos.DbRepo[*entities.Router]
	managedSvcRepo       repos.DbRepo[*entities.ManagedService]
	managedTemplatesPath string
	managedResRepo       repos.DbRepo[*entities.ManagedResource]
}

func fxDomain(
	workloadMessenger WorkloadMessenger,
	financeClient finance.FinanceClient,
	providerRepo repos.DbRepo[*entities.CloudProvider],
	consoleTemplate *template.Template,
	k8sYamlClient *k8s.YAMLClient,
	notifier rcn.ResourceChangeNotifier,
	deviceRepo repos.DbRepo[*entities.Device],
	clusterRepo repos.DbRepo[*entities.Cluster],
	projectRepo repos.DbRepo[*entities.Project],
	configRepo repos.DbRepo[*entities.Config],
	secretRepo repos.DbRepo[*entities.Secret],
	routerRepo repos.DbRepo[*entities.Router],
	regionRepo repos.DbRepo[*entities.EdgeRegion],
	appRepo repos.DbRepo[*entities.App],
	managedSvcRepo repos.DbRepo[*entities.ManagedService],
	managedResRepo repos.DbRepo[*entities.ManagedResource],
	instanceRepo repos.DbRepo[*entities.ResInstance],
	environmentRepo repos.DbRepo[*entities.Environment],
	msgP redpanda.Producer,
// env *Env,
	logger logging.Logger,
	ciClient ci.CIClient,
	iamClient iam.IAMClient,
	authClient auth.AuthClient,
	dnsClient kldns.DNSClient,
	changeNotifier rcn.ResourceChangeNotifier,
	jsEvalClient jseval.JSEvalClient,
	env *env.Env,
) Domain {
	return &domain{
		workloadMessenger:    workloadMessenger,
		financeClient:        financeClient,
		providerRepo:         providerRepo,
		consoleTemplate:      consoleTemplate,
		clusterRepo:          clusterRepo,
		k8sYamlClient:        k8sYamlClient,
		secretRepo:           secretRepo,
		projectRepo:          projectRepo,
		environmentRepo:      environmentRepo,
		appRepo:              appRepo,
		configRepo:           configRepo,
		routerRepo:           routerRepo,
		managedSvcRepo:       managedSvcRepo,
		managedResRepo:       managedResRepo,
		managedTemplatesPath: env.ManagedTemplatesPath,

		// instanceRepo:         instanceRepo,
		// changeNotifier:       changeNotifier,
		// notifier:             notifier,
		// ciClient:             ciClient,
		// authClient:           authClient,
		// iamClient:            iamClient,
		// deviceRepo:           deviceRepo,
		// messageProducer:      msgP,
		// logger:               logger,
		// inventoryPath:        env.InventoryPath,
		// jsEvalClient:         jsEvalClient,
		// regionRepo:           regionRepo,
		// dnsClient:            dnsClient,
		// clusterConfigsPath:   env.ClusterConfigsPath,
	}
}

var Module = fx.Module(
	"domain",
	// config.EnvFx[Env](),
	fx.Provide(fxClusterTemplate),
	fx.Provide(fxDomain),
)
