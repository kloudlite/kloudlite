package v1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

type ProvisionMode string

const (
	ProvisionModeOnDemand ProvisionMode = "on_demand"
	ProvisionModeSpot     ProvisionMode = "spot"
	ProvisionModeReserved ProvisionMode = "reserved"
)

type SpotSpecs struct {
	// +kubebuilder:validation:Minimum=0
	CpuMin int `json:"cpuMin"`
	// +kubebuilder:validation:Minimum=0
	CpuMax int `json:"cpuMax"`
	// +kubebuilder:validation:Minimum=0
	MemMin int `json:"memMin"`
	// +kubebuilder:validation:Minimum=0
	MemMax int `json:"memMax"`
}

type OnDemandSpecs struct {
	InstanceType string `json:"instanceType"`
}

type AWSNodeConfig struct {
	OnDemandSpecs *OnDemandSpecs `json:"onDemandSpecs,omitempty"`
	SpotSpecs     *SpotSpecs     `json:"spotSpecs,omitempty"`
	VPC           *string        `json:"vpc,omitempty"`
	Region        *string        `json:"region,omitempty"`
	IsGpu         bool           `json:"isGpu,omitempty"`
	// +kubebuilder:validation:Enum=on_demand;spot;reserved;
	ProvisionMode ProvisionMode `json:"provisionMode"`
	ImageId       *string       `json:"imageId,omitempty"`
}

type AwsNodePool struct {
	AMI                    string               `json:"ami"`
	AMISSHUsername         string               `json:"amiSSHUsername"`
	AvailabilityZone       *string              `json:"availabilityZone,omitempty"`
	InstanceType           string               `json:"instanceType"`
	NvidiaGpuEnabled       bool                 `json:"nvidiaGpuEnabled"`
	RootVolumeType         string               `json:"rootVolumeType"`
	RootVolumeSize         int                  `json:"rootVolumeSize"`
	IAMInstanceProfileRole *string              `json:"iamInstanceProfileRole,omitempty"`
	Nodes                  map[string]NodeProps `json:"nodes,omitempty"`
}

type AwsSpotNodePool struct {
	AMI                      string               `json:"ami"`
	AMISSHUsername           string               `json:"amiSSHUsername"`
	AvailabilityZone         *string              `json:"availabilityZone,omitempty"`
	NvidiaGpuEnabled         bool                 `json:"nvidiaGpuEnabled"`
	RootVolumeType           string               `json:"rootVolumeType"`
	RootVolumeSize           int                  `json:"rootVolumeSize"`
	IAMInstanceProfileRole   *string              `json:"iamInstanceProfileRole,omitempty"`
	SpotFleetTaggingRoleName string               `json:"spotFleetTaggingRoleName"`
	CpuNode                  *AwsSpotCpuNode      `json:"cpuNode,omitempty"`
	GpuNode                  *AwsSpotGpuNode      `json:"gpuNode,omitempty"`
	Nodes                    map[string]NodeProps `json:"nodes,omitempty"`
}

// NodePoolSpec defines the desired state of NodePool
// type NodePoolSpec struct {
// 	// +kubebuilder:validation:Minimum=0
// 	MaxCount int `json:"maxCount"`
// 	// +kubebuilder:validation:Minimum=0
// 	MinCount int `json:"minCount"`
// 	// +kubebuilder:validation:Minimum=0
// 	TargetCount int `json:"targetCount"`
//
// 	AWSNodeConfig *AWSNodeConfig `json:"awsNodeConfig,omitempty"`
//
// 	Taints []string          `json:"taints,omitempty"`
// 	Labels map[string]string `json:"labels,omitempty"`
// }

type AWSNodePoolConfig struct {
	// +kubebuilder:validation:Enum=normal;spot;
	PoolType string `json:"poolType"`

	NormalPool *AwsNodePool     `json:"normalPool,omitempty"`
	SpotPool   *AwsSpotNodePool `json:"spotPool,omitempty"`
}

type NodePoolSpec struct {
	// +kubebuilder:validation:Minimum=0
	MaxCount int `json:"maxCount"`
	// +kubebuilder:validation:Minimum=0
	MinCount int `json:"minCount"`
	// +kubebuilder:validation:Minimum=0
	TargetCount int `json:"targetCount"`

	// +kubebuilder:validation:Enum=aws;do;gcp;azure
	CloudProvider string `json:"cloudProvider"`

	AWS *AWSNodePoolConfig `json:"aws,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.nodepool-min-target-max",name=Min/Target/Max,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
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
		constants.GVKKey:          GroupVersion.WithKind("NodePool").String(),
		"nodepool-min-target-max": fmt.Sprintf("%d/%d/%d", n.Spec.MinCount, n.Spec.TargetCount, n.Spec.MaxCount),
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
