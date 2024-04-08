package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Peer struct {
	PublicKey string `json:"publicKey"`
	Endpoint  string `json:"endpoint"`
	Id        int    `json:"id"`

	AllowedIPs []string `json:"allowedIPs,omitempty"`
}

// ConnectionSpec defines the desired state of Connect
type ConnectionSpec struct {
	Id int `json:"id"`

	// PrivateKey *string `json:"privateKey,omitempty"`
	Interface *string `json:"interface,omitempty"`
	Nodeport  *int    `json:"nodeport,omitempty"`
	IpAddress *string `json:"ipAddress,omitempty"`
	DnsServer *string `json:"dnsServer,omitempty"`
	PublicKey *string `json:"publicKey,omitempty"`

	Peers []Peer `json:"peers,omitempty"`

	GatewayResources *corev1.ResourceRequirements `json:"gatewayResources,omitempty"`
	AgentsResources  *corev1.ResourceRequirements `json:"agentsResources,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ClusterConnection is the Schema for the connects API
type ClusterConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectionSpec `json:"spec,omitempty"`
	Status rApi.Status    `json:"status,omitempty" graphql:"noinput"`
}

func (d *ClusterConnection) EnsureGVK() {
	if d != nil {
		d.SetGroupVersionKind(GroupVersion.WithKind("Connection"))
	}
}

func (d *ClusterConnection) GetStatus() *rApi.Status {
	return &d.Status
}

func (d *ClusterConnection) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.WGDeviceNameKey: d.Name,
	}
}

func (d *ClusterConnection) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: GroupVersion.WithKind("Connection").String(),
	}
}

//+kubebuilder:object:root=true

// ClusterConnectionList contains a list of Connect
type ClusterConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterConnection{}, &ClusterConnectionList{})
}
