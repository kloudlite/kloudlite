package domain

import (
	"fmt"
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
	rcn "kloudlite.io/pkg/res-change-notifier"
	"math"
	"math/rand"
	"regexp"
	"strings"
)

type domain struct {
	deviceRepo           repos.DbRepo[*entities.Device]
	clusterRepo          repos.DbRepo[*entities.Cluster]
	projectRepo          repos.DbRepo[*entities.Project]
	configRepo           repos.DbRepo[*entities.Config]
	routerRepo           repos.DbRepo[*entities.Router]
	secretRepo           repos.DbRepo[*entities.Secret]
	messageProducer      redpanda.Producer
	messageTopic         string
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
	wgAccountRepo        repos.DbRepo[*entities.WGAccount]
	financeClient        finance.FinanceClient
	inventoryPath        string
}

func generateReadable(name string) string {
	compile := regexp.MustCompile(`[^0-9a-zA-Z:,/s]+`)
	allString := compile.ReplaceAllString(strings.ToLower(name), "")
	m := math.Min(10, float64(len(allString)))
	return fmt.Sprintf("%v_%v", allString[:int(m)], rand.Intn(9999))
}

type Env struct {
	KafkaInfraTopic      string `env:"KAFKA_INFRA_TOPIC" required:"true"`
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
	appRepo repos.DbRepo[*entities.App],
	managedSvcRepo repos.DbRepo[*entities.ManagedService],
	managedResRepo repos.DbRepo[*entities.ManagedResource],
	wgAccountRepo repos.DbRepo[*entities.WGAccount],
	msgP redpanda.Producer,
	env *Env,
	logger logging.Logger,
	workloadMessenger WorkloadMessenger,
	ciClient ci.CIClient,
	iamClient iam.IAMClient,
	authClient auth.AuthClient,
	financeClient finance.FinanceClient,
	changeNotifier rcn.ResourceChangeNotifier,
) Domain {
	return &domain{
		wgAccountRepo:        wgAccountRepo,
		changeNotifier:       changeNotifier,
		notifier:             notifier,
		ciClient:             ciClient,
		authClient:           authClient,
		iamClient:            iamClient,
		workloadMessenger:    workloadMessenger,
		deviceRepo:           deviceRepo,
		clusterRepo:          clusterRepo,
		projectRepo:          projectRepo,
		routerRepo:           routerRepo,
		secretRepo:           secretRepo,
		configRepo:           configRepo,
		appRepo:              appRepo,
		managedSvcRepo:       managedSvcRepo,
		managedResRepo:       managedResRepo,
		messageProducer:      msgP,
		messageTopic:         env.KafkaInfraTopic,
		managedTemplatesPath: env.ManagedTemplatesPath,
		logger:               logger,
		financeClient:        financeClient,
		inventoryPath:        env.InventoryPath,
	}
}

var Module = fx.Module(
	"domain",
	config.EnvFx[Env](),
	fx.Provide(fxDomain),
)
