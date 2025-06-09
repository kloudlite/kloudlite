package target

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TFCloudflareParams struct {
	Enabled  *bool  `json:"enabled"`
	ApiToken string `json:"api_token"`
	ZoneId   string `json:"zone_id"`
	Domain   string `json:"domain"`
}

type TFKloudliteParams struct {
	Release          string                 `json:"release"`
	InstallCRDs      bool                   `json:"install_crds"`
	InstallCsiDriver bool                   `json:"install_csi_driver"`
	InstallOperators bool                   `json:"install_operators"`
	InstallAgent     bool                   `json:"install_agent"`
	AgentVars        TFKloudliteAgentParams `json:"agent_vars"`
}

type TFKloudliteAgentParams struct {
	AccountName           string `json:"account_name"`
	ClusterName           string `json:"cluster_name"`
	ClusterToken          string `json:"cluster_token"`
	MessageOfficeGRPCAddr string `json:"message_office_grpc_addr"`
}

type AwsClusterTFValues struct {
	TrackerId     string `json:"tracker_id"`
	AwsRegion     string `json:"aws_region"`
	AwsAccessKey  string `json:"aws_access_key"`
	AwsSecretKey  string `json:"aws_secret_key"`
	AwsAssumeRole struct {
		Enabled    bool   `json:"enabled"`
		RoleARN    string `json:"role_arn"`
		ExternalID string `json:"external_id"`
	} `json:"aws_assume_role"`
	EnableNvidiaGPUSupport bool   `json:"enable_nvidia_gpu_support"`
	VpcID                  string `json:"vpc_id"`
	K3sMasters             struct {
		InstanceType     string `json:"instance_type"`
		NvidiaGpuEnabled bool   `json:"nvidia_gpu_enabled"`

		RootVolumeSize int    `json:"root_volume_size"`
		RootVolumeType string `json:"root_volume_type"`

		IAMInstanceProfile     string `json:"iam_instance_profile"`
		PublicDNSHost          string `json:"public_dns_host"`
		ClusterInternalDnsHost string `json:"cluster_internal_dns_host"`

		Cloudflare TFCloudflareParams `json:"cloudflare"`

		TaintMasterNodes bool `json:"taint_master_nodes"`

		BackupToS3 struct {
			Enabled bool `json:"enabled"`
		} `json:"backup_to_s3"`

		Nodes struct {
			Role             string       `json:"role"`
			AvailabilityZone string       `json:"availability_zone"`
			VpcSubnetId      string       `json:"vpc_subnet_id"`
			LastRecreatedAt  *metav1.Time `json:"last_recreated_at"`
			KloudliteRelease string       `json:"kloudlite_release"`
		} `json:"nodes"`
	} `json:"k3s_masters"`

	KloudliteParams TFKloudliteParams `json:"kloudlite_params"`
}

type TFGcpNode struct {
	AvailabilityZone string `json:"availability_zone"`
	K3SRole          string `json:"k3s_role"`
	KloudliteRelease string `json:"kloudlite_release"`
	BootvolumeType   string `json:"bootvolume_type"`
	BootvolumeSize   int    `json:"bootvolume_size"`
}

type GCPServiceAccount struct {
	Enabled bool     `json:"enabled"`
	Email   *string  `json:"email,omitempty"`
	Scopes  []string `json:"scopes,omitempty"`
}

type GcpClusterTFValues struct {
	GcpProjectId               string `json:"gcp_project_id"`
	GcpRegion                  string `json:"gcp_region"`
	GcpCredentialsJson         string `json:"gcp_credentials_json"`
	K3sSerivceCIDR             string `json:"k3s_service_cidr,omitempty"`
	NamePrefix                 string `json:"name_prefix"`
	ProvisionMode              string `json:"provision_mode"`
	Network                    string `json:"network"`
	UseAsLonghornStorageNodes  bool   `json:"use_as_longhorn_storage_nodes"`
	MachineType                string `json:"machine_type"`
	MachineState               string `json:"machine_state"`
	K3sDownloadURL             string `json:"k3s_download_url"`
	KloudliteRunnerDownloadURL string `json:"kloudlite_runner_download_url"`
	// "k3s_download_url": "https://github.com/kloudlite/infrastructure-as-code/releases/download/binaries/k3s",
	// "kloudlite_runner_download_url": "https://github.com/kloudlite/infrastructure-as-code/releases/download/binaries/runner-amd64",

	ServiceAccount       GCPServiceAccount    `json:"service_account"`
	Nodes                map[string]TFGcpNode `json:"nodes"`
	SaveSshKeyToPath     string               `json:"save_ssh_key_to_path,omitempty"`
	SaveKubeconfigToPath string               `json:"save_kubeconfig_to_path,omitempty"`
	PublicDnsHost        string               `json:"public_dns_host"`
	Cloudflare           TFCloudflareParams   `json:"cloudflare"`
	KloudliteParams      TFKloudliteParams    `json:"kloudlite_params"`
	Labels               map[string]string    `json:"labels"`
}
