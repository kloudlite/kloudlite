package v1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
)

// +kubebuilder:validation:Enum=ec2;spot;
type AWSPoolType string

const (
	AWSPoolTypeEC2  AWSPoolType = "ec2"
	AWSPoolTypeSpot AWSPoolType = "spot"
)

// +kubebuilder:validation:Enum=STANDARD;SPOT;
type GCPPoolType string

const (
	GCPPoolTypeStandard GCPPoolType = "STANDARD"
	GCPPoolTypeSpot     GCPPoolType = "SPOT"
)

type AwsEC2PoolConfig struct {
	InstanceType string               `json:"instanceType"`
	Nodes        map[string]NodeProps `json:"nodes,omitempty"`
}

type AwsSpotPoolConfig struct {
	SpotFleetTaggingRoleName string               `json:"spotFleetTaggingRoleName" graphql:"noinput"`
	CpuNode                  *AwsSpotCpuNode      `json:"cpuNode,omitempty"`
	GpuNode                  *AwsSpotGpuNode      `json:"gpuNode,omitempty"`
	Nodes                    map[string]NodeProps `json:"nodes,omitempty"`
}

type AWSNodePoolConfig struct {
	// ImageId          string `json:"imageId"`
	// ImageSSHUsername string `json:"imageSSHUsername"`

	VPCId       string `json:"vpcId" graphql:"noinput"`
	VPCSubnetID string `json:"vpcSubnetId" graphql:"noinput"`

	// AvailabilityZone AwsAZ `json:"availabilityZone"`
	AvailabilityZone string `json:"availabilityZone"`

	NvidiaGpuEnabled bool   `json:"nvidiaGpuEnabled"`
	RootVolumeType   string `json:"rootVolumeType" graphql:"noinput"`
	RootVolumeSize   int    `json:"rootVolumeSize" graphql:"noinput"`

	IAMInstanceProfileRole *string `json:"iamInstanceProfileRole,omitempty" graphql:"noinput"`

	PoolType AWSPoolType `json:"poolType"`

	EC2Pool  *AwsEC2PoolConfig  `json:"ec2Pool,omitempty"`
	SpotPool *AwsSpotPoolConfig `json:"spotPool,omitempty"`
}

type GCPNodePoolConfig struct {
	Region       string `json:"region" graphql:"noinput"`
	GCPProjectID string `json:"gcpProjectID" graphql:"noinput"`

	AvailabilityZone string `json:"availabilityZone"`

	// this secret's `.data` will be unmarshaled into type `GCPCredentials`
	Credentials ct.SecretRef `json:"credentials" graphql:"noinput"`

	PoolType GCPPoolType `json:"poolType"`

	MachineType    string `json:"machineType"`
	BootVolumeType string `json:"bootVolumeType" graphql:"noinput"`
	BootVolumeSize int    `json:"bootVolumeSize" graphql:"noinput"`

	Nodes map[string]NodeProps `json:"nodes,omitempty" graphql:"noinput"`
}

type Credentials struct {
	AccessKey ct.SecretKeyRef `json:"accessKey"`
	SecretKey ct.SecretKeyRef `json:"secretKey"`
}

type OperatorVars struct {
	JobName      string `json:"jobName"`
	JobNamespace string `json:"jobNamespace"`
}

type NodePoolSpec struct {
	// +kubebuilder:validation:Minimum=0
	MaxCount int `json:"maxCount"`
	// +kubebuilder:validation:Minimum=0
	MinCount int `json:"minCount"`

	NodeLabels map[string]string `json:"nodeLabels,omitempty"`
	NodeTaints []corev1.Taint    `json:"nodeTaints,omitempty"`

	CloudProvider ct.CloudProvider `json:"cloudProvider"`

	AWS *AWSNodePoolConfig `json:"aws,omitempty"`
	GCP *GCPNodePoolConfig `json:"gcp,omitempty"`
	// Azure *AzureNodePoolConfig `json:"azure,omitempty"`
	// DigitalOcean *DigitalOceanNodePoolConfig `json:"digitalocean,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.nodepool-min-max",name=Min/Max,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/nodepool\\.job-ref",name=JobRef,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// NodePool is the Schema for the nodepools API
type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodePoolSpec `json:"spec"`
	Status rApi.Status  `json:"status,omitempty" graphql:"noinput"`
}

func (n *NodePool) EnsureGVK() {
	if n != nil {
		n.SetGroupVersionKind(GroupVersion.WithKind("NodePool"))
	}
}

func (n *NodePool) GetStatus() *rApi.Status {
	return &n.Status
}

func (n *NodePool) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.NodePoolNameKey: n.Name,
	}
}

func (n *NodePool) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		"nodepool-min-max": fmt.Sprintf("%d/%d", n.Spec.MinCount, n.Spec.MaxCount),
	}
}

//+kubebuilder:object:root=true

// NodePoolList contains a list of NodePool
type NodePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodePool{}, &NodePoolList{})
}
