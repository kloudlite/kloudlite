package v1

import (
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

// NodePoolSpec defines the desired state of NodePool
type NodePoolSpec struct {
	// +kubebuilder:validation:Minimum=0
	MaxCount int `json:"maxCount"`
	// +kubebuilder:validation:Minimum=0
	MinCount int `json:"minCount"`
	// +kubebuilder:validation:Minimum=0
	TargetCount int `json:"targetCount"`

	AWSNodeConfig *AWSNodeConfig `json:"awsNodeConfig,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// NodePool is the Schema for the nodepools API
type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodePoolSpec `json:"spec"`
	Status rApi.Status  `json:"status,omitempty"`
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
		constants.NodePoolKey: n.Name,
	}
}

func (n *NodePool) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: GroupVersion.WithKind("NodePool").String(),
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
