package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

type ProvisionMode string

const (
	ProvisionModeOnDemand ProvisionMode = "on-demand"
	ProvisionModeSpot     ProvisionMode = "spot"
	ProvisionModeReserved ProvisionMode = "reserved"
)

type SpotSpecs struct {
	CpuMin int `json:"cpuMin"`
	CpuMax int `json:"cpuMax"`
	MemMin int `json:"memMin"`
	MemMax int `json:"memMax"`
}

type OnDemandSpecs struct {
	InstanceType string `json:"instanceType"`
}

type AWSNodeConfig struct {
	NodeName      *string        `json:"nodeName"`
	OnDemandSpecs *OnDemandSpecs `json:"onDemandSpecs"`
	SpotSpecs     *SpotSpecs     `json:"spotSpecs"`
	VPC           *string        `json:"vpc"`
	Region        *string        `json:"region"`
	ImageId       *string        `json:"imageId"`
	IsGpu         *bool          `json:"isGpu"`
	ProvisionMode ProvisionMode  `json:"provisionMode" enum:"on-demand;spot;reserved;"`
}

type CloudProvider string

const (
	CloudProviderAWS CloudProvider = "aws"
	CloudProviderGCP CloudProvider = "gcp"
)

// NodePoolSpec defines the desired state of NodePool
type NodePoolSpec struct {
	MaxCount    int `json:"maxCount"`
	MinCount    int `json:"minCount"`
	TargetCount int `json:"targetCount"`

	AWSNodeConfig *AWSNodeConfig `jons:"awsNodeConfig"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// NodePool is the Schema for the nodepools API
type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodePoolSpec `json:"spec,omitempty"`
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
	return map[string]string{}
}

func (n *NodePool) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: GroupVersion.WithKind("BYOC").String(),
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
