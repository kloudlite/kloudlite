package target

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

		Cloudflare struct {
			Enabled  bool   `json:"enabled"`
			ApiToken string `json:"api_token"`
			ZoneId   string `json:"zone_id"`
			Domain   string `json:"domain"`
		} `json:"cloudflare"`

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

  KloudliteParams struct{
    Release              string `json:"release"`
    InstallCRDs          bool   `json:"install_crds"`
    InstallCsiDriver     bool   `json:"install_csi_driver"`
    InstallOperators     bool   `json:"install_operators"`
    InstallAgent         bool   `json:"install_agent"`
    AgentVars struct {
      AccountName string `json:"account_name"`
      ClusterName string `json:"cluster_name"`
      ClusterToken string `json:"cluster_token"`
      MessageOfficeGRPCAddr string `json:"message_office_grpc_addr"`
    } `json:"agent_vars"`
  } `json:"kloudlite_params"`
}
