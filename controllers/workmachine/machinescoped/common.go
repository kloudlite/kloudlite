package machinescoped

import (
	"github.com/kloudlite/kloudlite/controllers/controllerconfig"
	controllersshared "github.com/kloudlite/kloudlite/controllers/shared"
	"github.com/kloudlite/kloudlite/pkg/errors"
	"github.com/kloudlite/kloudlite/pkg/operator-toolkit/kubectl"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Env struct {
	InstallationSecret string `env:"INSTALLATION_SECRET" required:"true"`

	HostManagerImage  string `env:"HOST_MANAGER_IMAGE" required:"true"`
	TunnelServerImage string `env:"TUNNEL_SERVER_IMAGE" required:"true"`
	CodeAnalyzerImage string `env:"CODE_ANALYZER_IMAGE" required:"true"`

	WorkMachineName string `env:"WORKMACHINE_NAME"`
	JWTSecret       string `env:"JWT_SECRET" required:"true"`
	HostedSubdomain string `env:"HOSTED_SUBDOMAIN" required:"true"`

	SnapshotRegistryEndpoint string `env:"SNAPSHOT_REGISTRY_ENDPOINT" default:"image-registry.kloudlite.svc.cluster.local:5000"`
	SnapshotRegistryPrefix   string `env:"SNAPSHOT_REGISTRY_PREFIX" default:"snapshots"`
	SnapshotRegistryInsecure string `env:"SNAPSHOT_REGISTRY_INSECURE" default:"true"`

	NodeName string `env:"NODE_NAME"`
}

type MachineScopedReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	YAMLClient kubectl.YAMLClient
	env        Env
	Cfg        *controllerconfig.ControllerConfig
}

const (
	WorkMachineFinalizerName = "workmachine.machines.kloudlite.io/cleanup"
	SSHUserName              = "kloudlite"
	wmIngressControllerImage = "ghcr.io/kloudlite/kloudlite/wm-ingress-controller:development"
)

func (r *MachineScopedReconciler) initSharedRuntimeState() error {
	if podDeletionTracker == nil {
		logger, err := zap.NewProduction()
		if err != nil {
			return errors.Wrap("failed to create logger for pod deletion tracker", err)
		}
		podDeletionTracker = controllersshared.NewPodDeletionTracker(logger)
	}
	return nil
}
