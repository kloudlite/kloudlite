package env

import (
	"io"
	"os"

	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
	"sigs.k8s.io/yaml"
)

type Env struct {
	InfraDbUri  string `env:"MONGO_DB_URI" required:"true"`
	InfraDbName string `env:"MONGO_DB_NAME" required:"true"`

	HttpPort     uint16 `env:"HTTP_PORT" required:"true"`
	GrpcPort     uint16 `env:"GRPC_PORT" required:"true"`
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	KloudliteDNSSuffix string `env:"KLOUDLITE_DNS_SUFFIX" required:"true"`

	NatsURL    string `env:"NATS_URL" required:"true"`
	NatsStream string `env:"NATS_STREAM" required:"true"`

	AccountCookieName       string `env:"ACCOUNT_COOKIE_NAME" required:"true"`
	ProviderSecretNamespace string `env:"PROVIDER_SECRET_NAMESPACE" required:"true"`

	IAMGrpcAddr      string `env:"IAM_GRPC_ADDR" required:"true"`
	AccountsGrpcAddr string `env:"ACCOUNTS_GRPC_ADDR" required:"true"`
	ConsoleGrpcAddr  string `env:"CONSOLE_GRPC_ADDR" required:"true"`

	MessageOfficeInternalGrpcAddr string `env:"MESSAGE_OFFICE_INTERNAL_GRPC_ADDR" required:"true"`
	MessageOfficeExternalGrpcAddr string `env:"MESSAGE_OFFICE_EXTERNAL_GRPC_ADDR" required:"true"`

	AWSCfParamTrustedARN           string `env:"AWS_CF_PARAM_TRUSTED_ARN" required:"true"`
	AWSCfStackNamePrefix           string `env:"AWS_CF_STACK_NAME_PREFIX" required:"true"`
	AWSCfRoleNamePrefix            string `env:"AWS_CF_ROLE_NAME_PREFIX" required:"true"`
	AWSCfInstanceProfileNamePrefix string `env:"AWS_CF_INSTANCE_PROFILE_NAME_PREFIX" required:"true"`
	AWSCfStackS3URL                string `env:"AWS_CF_STACK_S3_URL" required:"true"`

	AWSAccessKey string `env:"AWS_ACCESS_KEY" required:"true"`
	AWSSecretKey string `env:"AWS_SECRET_KEY" required:"true"`

	PublicDNSHostSuffix string `env:"PUBLIC_DNS_HOST_SUFFIX" required:"true"`
	SessionKVBucket     string `env:"SESSION_KV_BUCKET" required:"true"`

	MsvcTemplateFilePath string `env:"MSVC_TEMPLATE_FILE_PATH" required:"true"`

	KloudliteRelease string `env:"KLOUDLITE_RELEASE" required:"true"`

	// READ more @ https://tailscale.com/kb/1015/100.x-addresses
	BaseCIDR string `env:"BASE_CIDR" default:"100.64.0.0/10"`
	// 18, as for 16K (2**14) IPs per cluster
	AllocatableCIDRSuffix int `env:"ALLOCATABLE_CIDR_SUFFIX" default:"18"`

	// 20, as for  4K (2**12) IPs per cluster, reserved for k8s services
	AllocatableSvcCIDRSuffix int `env:"ALLOCATABLE_SVC_CIDR_SUFFIX" default:"20"`
	// ClusterOffset = 5, reserving 5 * 8K IPs for wireguard devices and other devices, that are not Clusters
	ClustersOffset int `env:"CLUSTERS_OFFSET" default:"5"`

	IsDev              bool
	KubernetesApiProxy string `env:"KUBERNETES_API_PROXY"`

	GlobalVPNKubeReverseProxyImage      string `env:"GLOBAL_VPN_KUBE_REVERSE_PROXY_IMAGE" required:"true"`
	GlobalVPNKubeReverseProxyAuthzToken string `env:"GLOBAL_VPN_KUBE_REVERSE_PROXY_AUTHZ_TOKEN" required:"true"`

	KloudliteGlobalVPNDeviceHost string `env:"KLOUDLITE_GLOBAL_VPN_DEVICE_HOST" required:"true"`

	AvailableKloudliteRegionsConfig string `env:"AVAILABLE_KLOUDLITE_REGIONS_CONFIG" required:"false"`
	AvailableKloudliteRegions       map[string]AvailableKloudliteRegion
}

type AvailableKloudliteRegion struct {
	ID            string `json:"id"`
	DisplayName   string `json:"displayName"`
	Region        string `json:"region"`
	CloudProvider string `json:"cloudProvider"`
	Kubeconfig    string `json:"kubeconfig"`
	PublicDNSHost string `json:"publicDNSHost"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, errors.NewE(err)
	}
	if ev.AvailableKloudliteRegionsConfig != "" {
		f, err := os.Open(ev.AvailableKloudliteRegionsConfig)
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		var regions []AvailableKloudliteRegion
		if err := yaml.Unmarshal(b, &regions); err != nil {
			return nil, err
		}
		ev.AvailableKloudliteRegions = make(map[string]AvailableKloudliteRegion, len(regions))
		for i := range regions {
			ev.AvailableKloudliteRegions[regions[i].ID] = regions[i]
		}
	}
	return &ev, nil
}
