package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Peer struct {
	PublicKey string `json:"publicKey"`
	Endpoint  string `json:"endpoint"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`

	ClusterName string `json:"clusterName,omitempty"`
	DeviceName  string `json:"deviceName,omitempty"`

	AllowedIPs []string `json:"allowedIPs,omitempty"`
}

type WgParams struct {
	WgPrivateKey string `json:"wg_private_key"`
	WgPublicKey  string `json:"wg_public_key"`

	IP string `json:"ip"`

	DNSServer *string `json:"dnsServer"`

	PublicGatewayHosts *string `json:"publicGatewayHosts,omitempty"`
	PublicGatewayPort  *string `json:"publicGatewayPort,omitempty"`

	VirtualCidr string `json:"virtualCidr"`
}

// ConnectionSpec defines the desired state of Connect
type GlobVPNSpec struct {
	// This secret is unmarshalled into WgParams
	WgRef ct.SecretRef `json:"wg"`

	WgInterface *string `json:"wgInterface"`

	Peers []Peer `json:"peers,omitempty"`

	GatewayResources *corev1.ResourceRequirements `json:"gatewayResources,omitempty"`
	AgentsResources  *corev1.ResourceRequirements `json:"agentsResources,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
// GlobalVPN is the Schema for the connects API
type GlobalVPN struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlobVPNSpec `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty" graphql:"noinput"`
}

func (d *GlobalVPN) EnsureGVK() {
	if d != nil {
		d.SetGroupVersionKind(GroupVersion.WithKind("Connection"))
	}
}

func (d *GlobalVPN) GetStatus() *rApi.Status {
	return &d.Status
}

func (d *GlobalVPN) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.WGDeviceNameKey: d.Name,
	}
}

func (d *GlobalVPN) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: GroupVersion.WithKind("Connection").String(),
	}
}

//+kubebuilder:object:root=true

// GlobalVPNList contains a list of Connect
type GlobalVPNList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalVPN `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GlobalVPN{}, &GlobalVPNList{})
}
