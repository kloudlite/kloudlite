package platformscoped

import (
	"os"

	"github.com/kloudlite/kloudlite/controllers/controllerconfig"
	"github.com/kloudlite/kloudlite/controllers/workmachine/cloud"
	"github.com/kloudlite/kloudlite/pkg/errors"
	"github.com/kloudlite/kloudlite/pkg/operator-toolkit/kubectl"
	v1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Env struct {
	KloudliteInstallationID string           `env:"INSTALLATION_KEY" required:"true"`
	K3sVersion              string           `env:"K3S_VERSION" required:"true"`
	K3sServerURL            string           `env:"K3S_SERVER_URL" required:"true"`
	K3sAgentToken           string           `env:"K3S_AGENT_TOKEN" required:"true"`
	CloudProvider           v1.CloudProvider `env:"CLOUD_PROVIDER" required:"true"`
	HostedSubdomain         string           `env:"HOSTED_SUBDOMAIN" required:"true"`

	WorkMachineManagerImage string `env:"WORKMACHINE_MANAGER_IMAGE" default:"ghcr.io/kloudlite/kloudlite/workmachine-manager:development"`
	PodNamespace            string `env:"POD_NAMESPACE" default:"kloudlite"`
}

type PlatformScopedReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	YAMLClient kubectl.YAMLClient

	env              Env
	cloudProviderAPI cloud.Provider
	usageReporter    *UsageReporter

	Cfg *controllerconfig.ControllerConfig
}

const (
	NodeLabelPublicIP  = "kloudlite.io/public-ip"
	NodeLabelPrivateIP = "kloudlite.io/private-ip"
)

func (r *PlatformScopedReconciler) initSharedRuntimeState() error {
	consoleBaseURL := os.Getenv("CONSOLE_BASE_URL")
	if consoleBaseURL == "" {
		consoleBaseURL = "https://console.kloudlite.io"
	}
	logger, err := zap.NewProduction()
	if err != nil {
		return errors.Wrap("failed to create logger for usage reporter", err)
	}
	r.usageReporter = NewUsageReporter(consoleBaseURL, r.env.KloudliteInstallationID, logger)
	return nil
}
