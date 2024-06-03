package env

import "github.com/codingconcepts/env"

type Env struct {
	IsDev bool

	GatewayWGPublicKey  string `env:"GATEWAY_WG_PUBLIC_KEY" required:"true"`
	GatewayWGPrivateKey string `env:"GATEWAY_WG_PRIVATE_KEY" required:"true"`
	GatewayWGEndpoint   string `env:"GATEWAY_WG_ENDPOINT" required:"true"`
	GatewayGlobalIP     string `env:"GATEWAY_GLOBAL_IP" required:"true"`

	ClusterCIDR string `env:"CLUSTER_CIDR" required:"true"`
	ServiceCIDR string `env:"SERVICE_CIDR" required:"true"`

	IPManagerConfigName      string `env:"IP_MANAGER_CONFIG_NAME" required:"true"`
	IPManagerConfigNamespace string `env:"IP_MANAGER_CONFIG_NAMESPACE" required:"true"`

	PodAllowedIPs string `env:"POD_ALLOWED_IPS" required:"true"`

	NginxStreamsDir    string `env:"NGINX_STREAMS_DIR" default:"/etc/nginx/streams.d"`
	WireguardConfigDir string `env:"WIREGUARD_CONFIG_DIR" default:"/etc/wireguard"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
