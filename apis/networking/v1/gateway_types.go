package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Peer struct {
	PublicKey      string  `json:"publicKey"`
	PublicEndpoint *string `json:"publicEndpoint,omitempty"`
	IP             string  `json:"ip"`

	DNSSuffix  *string  `json:"dnsSuffix,omitempty"`
	AllowedIPs []string `json:"allowedIPs,omitempty"`
}

type GatewayLoadBalancer struct {
	Hosts []string `json:"hosts"`
	Port  int32    `json:"port"`
}

// GatewaySpec defines the desired state of Gateway
type GatewaySpec struct {
	AdminNamespace string `json:"adminNamespace,omitempty"`

	GlobalIP string `json:"globalIP"`

	ClusterCIDR string `json:"clusterCIDR"`
	SvcCIDR     string `json:"svcCIDR"`

	DNSSuffix string `json:"dnsSuffix"`

	Peers []Peer `json:"peers,omitempty"`

	// it will be filled by resource controller
	LoadBalancer *GatewayLoadBalancer `json:"loadBalancer,omitempty"`

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
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
//+kubebuilder:printcolumn:JSONPath=".spec.globalIP",name=GlobalIP,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.clusterCIDR",name=ClusterCIDR,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.svcCIDR",name=ServiceCIDR,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string

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
