package domain

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"

	kldns "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/dns"

	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/jseval"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/kubeapi"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
	rcn "kloudlite.io/pkg/res-change-notifier"
)

type domain struct {
	kubeCli         *kubeapi.Client
	deviceRepo      repos.DbRepo[*entities.Device]
	clusterRepo     repos.DbRepo[*entities.Cluster]
	projectRepo     repos.DbRepo[*entities.Project]
	configRepo      repos.DbRepo[*entities.Config]
	routerRepo      repos.DbRepo[*entities.Router]
	secretRepo      repos.DbRepo[*entities.Secret]
	regionRepo      repos.DbRepo[*entities.EdgeRegion]
	messageProducer redpanda.Producer
	// messageTopic         string
	logger               logging.Logger
	managedSvcRepo       repos.DbRepo[*entities.ManagedService]
	managedResRepo       repos.DbRepo[*entities.ManagedResource]
	appRepo              repos.DbRepo[*entities.App]
	managedTemplatesPath string
	workloadMessenger    WorkloadMessenger
	ciClient             ci.CIClient
	imageRepoUrlPrefix   string
	notifier             rcn.ResourceChangeNotifier
	iamClient            iam.IAMClient
	authClient           auth.AuthClient
	changeNotifier       rcn.ResourceChangeNotifier
	financeClient        finance.FinanceClient
	inventoryPath        string
	jsEvalClient         jseval.JSEvalClient
	providerRepo         repos.DbRepo[*entities.CloudProvider]
	dnsClient            kldns.DNSClient
}

func generateReadable(name string) string {
	compile := regexp.MustCompile(`[^0-9a-zA-Z:,/s]+`)
	allString := compile.ReplaceAllString(strings.ToLower(name), "")
	m := math.Min(10, float64(len(allString)))
	return fmt.Sprintf("%v_%v", allString[:int(m)], rand.Intn(9999))
}

type Env struct {
	// KafkaInfraTopic      string `env:"KAFKA_INFRA_TOPIC" required:"true"`
	ManagedTemplatesPath string `env:"MANAGED_TEMPLATES_PATH" required:"true"`
	InventoryPath        string `env:"INVENTORY_PATH" required:"true"`
}

func fxDomain(
	notifier rcn.ResourceChangeNotifier,
	deviceRepo repos.DbRepo[*entities.Device],
	clusterRepo repos.DbRepo[*entities.Cluster],
	projectRepo repos.DbRepo[*entities.Project],
	configRepo repos.DbRepo[*entities.Config],
	secretRepo repos.DbRepo[*entities.Secret],
	routerRepo repos.DbRepo[*entities.Router],
	regionRepo repos.DbRepo[*entities.EdgeRegion],
	providerRepo repos.DbRepo[*entities.CloudProvider],
	appRepo repos.DbRepo[*entities.App],
	managedSvcRepo repos.DbRepo[*entities.ManagedService],
	managedResRepo repos.DbRepo[*entities.ManagedResource],
	msgP redpanda.Producer,
	env *Env,
	logger logging.Logger,
	workloadMessenger WorkloadMessenger,
	ciClient ci.CIClient,
	iamClient iam.IAMClient,
	authClient auth.AuthClient,
	financeClient finance.FinanceClient,
	dnsClient kldns.DNSClient,
	changeNotifier rcn.ResourceChangeNotifier,
	jsEvalClient jseval.JSEvalClient,
	kubecli *kubeapi.Client,
) Domain {
	return &domain{
		kubeCli:           kubecli,
		providerRepo:      providerRepo,
		changeNotifier:    changeNotifier,
		notifier:          notifier,
		ciClient:          ciClient,
		authClient:        authClient,
		iamClient:         iamClient,
		workloadMessenger: workloadMessenger,
		deviceRepo:        deviceRepo,
		clusterRepo:       clusterRepo,
		projectRepo:       projectRepo,
		routerRepo:        routerRepo,
		secretRepo:        secretRepo,
		configRepo:        configRepo,
		appRepo:           appRepo,
		managedSvcRepo:    managedSvcRepo,
		managedResRepo:    managedResRepo,
		messageProducer:   msgP,
		// messageTopic:         env.KafkaInfraTopic,
		managedTemplatesPath: env.ManagedTemplatesPath,
		logger:               logger,
		financeClient:        financeClient,
		inventoryPath:        env.InventoryPath,
		jsEvalClient:         jsEvalClient,
		regionRepo:           regionRepo,
		dnsClient:            dnsClient,
	}
}

var Module = fx.Module(
	"domain",
	config.EnvFx[Env](),
	fx.Provide(fxDomain),
)
