package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

type NodeType string

const (
	NodeTypeWorker  NodeType = "worker"
	NodeTypeMaster  NodeType = "master"
	NodeTypeCluster NodeType = "cluster"
)

type NodeSpec struct {
	NodePoolName string `json:"nodePoolName"`
	// +kubebuilder:validation:Enum=worker;master;cluster
	NodeType NodeType `json:"nodeType"`
	Taints   []string `json:"taints,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Node is the Schema for the nodes API
type Node struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeSpec    `json:"spec"`
	Status rApi.Status `json:"status,omitempty"`
}

func (n *Node) EnsureGVK() {
	if n != nil {
		n.SetGroupVersionKind(GroupVersion.WithKind("Node"))
	}
}

func (n *Node) GetStatus() *rApi.Status {
	return &n.Status
}

func (n *Node) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.NodePoolKey: n.Spec.NodePoolName,
		constants.NodeNameKey: n.Name,
	}
}

func (n *Node) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: GroupVersion.WithKind("Node").String(),
	}
}

//+kubebuilder:object:root=true

// NodeList contains a list of Node
type NodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Node `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Node{}, &NodeList{})
}
