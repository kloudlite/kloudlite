package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

// NodePoolSpec defines the desired state of NodePool
type NodePoolSpec struct {
	MaxCount    int `json:"maxCount"`
	MinCount    int `json:"minCount"`
	TargetCount int `json:"targetCount"`

	// aws -> CloudProvider
	NodeConfig string `json:"nodeConfig"`

	// IsStateful bool `json:"isStateful,omitempty"`

	// aws secrets
	// account name
}

// node auto scaler -> del, create
// 4

// clusters.kloudlite.io/node
/*
provier secret
accountId
node name
node type(cluster, secondary-master, worker)
node config

*/

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
