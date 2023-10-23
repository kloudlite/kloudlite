package v1

import (
	common_types "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// type NodeConfig struct {
// 	InstanceType     string `json:"instanceType"`
// 	AvailabilityZone string `json:"availabilityZone"`
// 	RootVolumeSize   int    `json:"rootVolumeSize"`
// 	// +kubebuilder:validation:Enum=primary-master;secondary-master;agent;
// 	Role            string `json:"role"`
// 	IsNvidiaGpuNode *bool  `json:"isNvidiaGpuNode"`
// }

// type NvidiaGpuOpts struct {
// 	Enabled       bool     `json:"enabled"`
// 	InstanceTypes []string `json:"instanceTypes,omitempty"`
// }
//
// type SpotNodeConfig struct {
// 	VCpu           common_types.MinMaxInt `json:"vCpu"`
// 	MemPerVCpu     common_types.MinMaxInt `json:"memPerVCpu"`
// 	RootVolumeSize int                    `json:"rootVolumeSize"`
// 	NvidiaGpuOpts  *NvidiaGpuOpts         `json:"nvidiaGpuOpts,omitempty"`
// }

// type AWSSpotSettings struct {
// 	Enabled                  bool   `json:"enabled"`
// 	SpotFleetTaggingRoleName string `json:"spotFleetTaggingRoleName"`
// }

type AwsSpotCpuNode struct {
	VCpu          common_types.MinMaxFloat `json:"vcpu"`
	MemoryPerVCpu common_types.MinMaxFloat `json:"memoryPerVcpu,omitempty"`
}

type AwsSpotGpuNode struct {
	InstanceTypes []string `json:"instanceTypes"`
}

type MasterNodeProps struct {
	// +kubebuilder:validation:Enum=primary-master;secondary-master;
	Role             string `json:"role"`
	AvaialbilityZone string `json:"availabilityZone"`
	NodeProps        `json:",inline"`
}

type NodeProps struct {
	LastRecreatedAt *metav1.Time `json:"lastRecreatedAt,omitempty"`
}

type AWSK3sMastersConfig struct {
	AMI                    string                     `json:"ami"`
	AMISSHUsername         string                     `json:"amiSSHUsername"`
	InstanceType           string                     `json:"instanceType"`
	NvidiaGpuEnabled       bool                       `json:"nvidiaGpuEnabled"`
	RootVolumeType         string                     `json:"rootVolumeType"`
	RootVolumeSize         int                        `json:"rootVolumeSize"`
	IAMInstanceProfileRole *string                    `json:"iamInstanceProfileRole,omitempty"`
	PublicDNSHost          *string                    `json:"publicDnsHost,omitempty"`
	ClusterInternalDnsHost *string                    `json:"clusterInternalDnsHost,omitempty"`
	CloudflareEnabled      *bool                      `json:"cloudflareEnabled,omitempty"`
	TaintMasterNodes       bool                       `json:"taintMasterNodes"`
	BackupToS3Enabled      bool                       `json:"backupToS3Enabled"`
	Nodes                  map[string]MasterNodeProps `json:"nodes,omitempty"`
}

type AWSClusterConfig struct {
	Region string `json:"region"`
	// AMI    string `json:"ami"`

	// IAMInstanceProfileRole *string `json:"iamInstanceProfileRole,omitempty"`
	// EC2NodesConfig         map[string]NodeConfig `json:"ec2NodesConfig,omitempty"`
	K3sMasters AWSK3sMastersConfig `json:"k3sMasters,omitempty"`

	NodePools     map[string]AwsNodePool     `json:"nodePools,omitempty"`
	SpotNodePools map[string]AwsSpotNodePool `json:"spotNodePools,omitempty"`

	// SpotSettings    *AWSSpotSettings          `json:"spotSettings,omitempty"`
	// SpotNodesConfig map[string]SpotNodeConfig `json:"spotNodesConfig,omitempty"`
}

type DigitalOceanConfig struct{}

type AzureConfig struct{}

type GCPConfig struct{}

// ClusterSpec defines the desired state of Cluster
// For now considered basis on AWS Specific
type ClusterSpec struct {
	AccountName string  `json:"accountName"`
	AccountId   *string `json:"accountId,omitempty"`

	ClusterTokenRef common_types.SecretKeyRef `json:"clusterTokenRef,omitempty"`

	DNSHostName *string `json:"dnsHostName,omitempty"`

	CredentialsRef common_types.SecretRef `json:"credentialsRef"`

	// +kubebuilder:validation:Enum=dev;HA
	AvailabilityMode string `json:"availabilityMode"`

	// +kubebuilder:validation:Enum=aws;do;gcp;azure
	CloudProvider string `json:"cloudProvider"`

	AWS          *AWSClusterConfig   `json:"aws,omitempty"`
	DigitalOcean *DigitalOceanConfig `json:"do,omitempty"`
	GCP          *GCPConfig          `json:"gcp,omitempty"`
	Azure        *AzureConfig        `json:"azure,omitempty"`

	// // +kubebuilder:validation:default=false
	// DisableSSH bool `json:"disableSSH,omitempty"`

	MessageQueueTopicName *string `json:"messageQueueTopicName,omitempty"`

	// NodeIps []string `json:"nodeIps,omitempty"`
	// VPC     *string  `json:"vpc,omitempty"`

	// AgentHelmValues     *common_types.SecretKeyRef `json:"agentHelmValuesRef,omitempty"`
	// OperatorsHelmValues *common_types.SecretKeyRef `json:"operatorsHelmValuesRef,omitempty"`

	KloudliteRelease string `json:"kloudliteRelease"`
}

// type KloudliteParams struct {
// 	Release          string `json:"release,omitempty"`
// 	InstallCRDs      bool   `json:"installCRDs,omitempty"`
// 	InstallCSIDriver bool   `json:"installCSIDriver,omitempty"`
// 	InstallOperators bool   `json:"installOperators,omitempty"`
// 	InstallAgent     bool   `json:"installAgent,omitempty"`
// }

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.accountName",name=AccountName,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.messageQueueTopicName",name=QTopic,type=string
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (b *Cluster) EnsureGVK() {
	if b != nil {
		b.SetGroupVersionKind(GroupVersion.WithKind("Cluster"))
	}
}

func (b *Cluster) GetStatus() *rApi.Status {
	return &b.Status
}

func (b *Cluster) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (b *Cluster) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: GroupVersion.WithKind("Cluster").String(),
	}
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
