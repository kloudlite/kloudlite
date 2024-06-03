package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GatewaySpec defines the desired state of Gateway
type GatewaySpec struct {
	GlobalIP string `json:"globalIP"`

	ClusterCIDR string `json:"clusterCIDR"`
	SvcCIDR     string `json:"svcCIDR"`

	// secret's data will be unmarshalled into WireguardKeys
	WireguardKeysRef ct.SecretRef `json:"wireguardKeysRef,omitempty"`
}

type WireguardKeys struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:JSONPath=".spec.clusterCIDR",name=ClusterCIDR,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.svcCIDR",name=ServiceCIDR,type=string

// Gateway is the Schema for the gateways API
type Gateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewaySpec `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (p *Gateway) EnsureGVK() {
	if p != nil {
		p.SetGroupVersionKind(GroupVersion.WithKind("Gateway"))
	}
}

func (p *Gateway) GetStatus() *rApi.Status {
	return &p.Status
}

func (p *Gateway) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (p *Gateway) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// GatewayList contains a list of Gateway
type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gateway{}, &GatewayList{})
}
