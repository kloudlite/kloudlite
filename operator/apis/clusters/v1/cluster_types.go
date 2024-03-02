package v1

import (
	common_types "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AwsSpotCpuNode struct {
	VCpu          common_types.MinMaxFloat `json:"vcpu"`
	MemoryPerVCpu common_types.MinMaxFloat `json:"memoryPerVcpu,omitempty"`
}

type AwsSpotGpuNode struct {
	InstanceTypes []string `json:"instanceTypes"`
}

type MasterNodeProps struct {
	// +kubebuilder:validation:Enum=primary-master;secondary-master;
	Role string `json:"role"`
	// AvailabilityZone AwsAZ  `json:"availabilityZone"`
	AvailabilityZone string `json:"availabilityZone"`
	KloudliteRelease string `json:"kloudliteRelease"`
	NodeProps        `json:",inline"`
}

type NodeProps struct {
	LastRecreatedAt *metav1.Time `json:"lastRecreatedAt,omitempty"`
}

type AWSK3sMastersConfig struct {
	InstanceType           string                     `json:"instanceType"`
	NvidiaGpuEnabled       bool                       `json:"nvidiaGpuEnabled"`
	RootVolumeType         string                     `json:"rootVolumeType" graphql:"noinput"`
	RootVolumeSize         int                        `json:"rootVolumeSize" graphql:"noinput"`
	IAMInstanceProfileRole *string                    `json:"iamInstanceProfileRole,omitempty" graphql:"noinput"`
	Nodes                  map[string]MasterNodeProps `json:"nodes,omitempty" graphql:"noinput"`
}

type CloudProviderCredentialKeys struct {
	KeyAWSAccountId              string `json:"keyAWSAccountId"`
	KeyAWSAssumeRoleExternalID   string `json:"keyAWSAssumeRoleExternalID"`
	KeyAWSAssumeRoleRoleARN      string `json:"keyAWSAssumeRoleRoleARN"`
	KeyAWSIAMInstanceProfileRole string `json:"keyIAMInstanceProfileRole"`
	KeyAccessKey                 string `json:"keyAccessKey"`
	KeySecretKey                 string `json:"keySecretKey"`
}

type AwsSubnetWithID struct {
	AvailabilityZone string `json:"availabilityZone"`
	ID               string `json:"id"`
}

type AwsVPCParams struct {
	ID            string            `json:"id"`
	PublicSubnets []AwsSubnetWithID `json:"publicSubnets"`
}

type AWSClusterConfig struct {
	VPC *AwsVPCParams `json:"vpc,omitempty" graphql:"noinput"`

	// Region     AwsRegion           `json:"region"`
	Region     string              `json:"region"`
	K3sMasters AWSK3sMastersConfig `json:"k3sMasters,omitempty"`

	NodePools     map[string]AwsEC2PoolConfig  `json:"nodePools,omitempty" graphql:"noinput"`
	SpotNodePools map[string]AwsSpotPoolConfig `json:"spotNodePools,omitempty" graphql:"noinput"`
}

func (avp *AwsVPCParams) GetSubnetId(az string) string {
	if avp == nil {
		return ""
	}

	for _, v := range avp.PublicSubnets {
		if v.AvailabilityZone == az {
			return v.ID
		}
	}

	return ""
}

type DigitalOceanConfig struct{}

type AzureConfig struct{}

type GCPConfig struct{}

type ClusterOutput struct {
	JobName      string `json:"jobName"`
	JobNamespace string `json:"jobNamespace"`

	SecretName string `json:"secretName"`

	KeyKubeconfig          string `json:"keyKubeconfig"`
	KeyK3sServerJoinToken  string `json:"keyK3sServerJoinToken"`
	KeyK3sAgentJoinToken   string `json:"keyK3sAgentJoinToken"`
	KeyAWSVPCId            string `json:"keyAWSVPCId,omitempty"`
	KeyAWSVPCPublicSubnets string `json:"keyAWSVPCPublicSubnets,omitempty"`
}

// ClusterSpec defines the desired state of Cluster
// For now considered basis on AWS Specific
type ClusterSpec struct {
	AccountName     string                       `json:"accountName" graphql:"noinput"`
	AccountId       string                       `json:"accountId" graphql:"noinput"`
	ClusterTokenRef common_types.SecretKeyRef    `json:"clusterTokenRef,omitempty" graphql:"noinput"`
	CredentialsRef  common_types.SecretRef       `json:"credentialsRef"`
	CredentialKeys  *CloudProviderCredentialKeys `json:"credentialKeys,omitempty" graphql:"noinput"`

	// +kubebuilder:validation:Enum=dev;HA
	AvailabilityMode string `json:"availabilityMode" graphql:"enum=dev;HA"`

	TaintMasterNodes       bool    `json:"taintMasterNodes" graphql:"noinput"`
	BackupToS3Enabled      bool    `json:"backupToS3Enabled" graphql:"noinput"`
	PublicDNSHost          string  `json:"publicDNSHost" graphql:"noinput"`
	ClusterInternalDnsHost *string `json:"clusterInternalDnsHost,omitempty" graphql:"noinput"`
	CloudflareEnabled      *bool   `json:"cloudflareEnabled,omitempty"`

	// +kubebuilder:validation:Enum=aws;do;gcp;azure
	CloudProvider common_types.CloudProvider `json:"cloudProvider"`

	AWS *AWSClusterConfig `json:"aws,omitempty"`

	MessageQueueTopicName string `json:"messageQueueTopicName" graphql:"noinput"`
	KloudliteRelease      string `json:"kloudliteRelease" graphql:"noinput"`

	Output *ClusterOutput `json:"output,omitempty" graphql:"noinput"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.accountName",name=AccountName,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/cluster\\.job-ref",name=Job,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   ClusterSpec `json:"spec"`
	Status rApi.Status `json:"status,omitempty" graphql:"noinput"`
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
	return map[string]string{
		constants.AccountNameKey: b.Spec.AccountName,
		constants.RegionKey: func() string {
			if b.Spec.AWS != nil {
				return string(b.Spec.AWS.Region)
			}
			return ""
		}(),
	}
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
