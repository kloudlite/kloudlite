package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KloudliteDnsApi struct {
	PublicHttpUri  string          `json:"publicHttpUri"`
	BasicAuthCreds SecretReference `json:"basicAuthCreds"`
}

// ManagedClusterSpec defines the desired state of ManagedCluster
type ManagedClusterSpec struct {
	Domain         *string         `json:"domain,omitempty"`
	KloudliteCreds SecretReference `json:"kloudliteCreds,omitempty"`
}

type KloudliteCreds struct {
	DnsApiEndpoint string `json:"dnsApiEndpoint,omitempty" validate:"required"`
	DnsApiUsername string `json:"dnsApiUsername" validate:"required"`
	DnsApiPassword string `json:"dnsApiPassword" validate:"required"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ManagedCluster is the Schema for the managedclusters API
type ManagedCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedClusterSpec `json:"spec,omitempty"`
	Status rApi.Status        `json:"status,omitempty"`
}

func (mc *ManagedCluster) GetStatus() *rApi.Status {
	return &mc.Status
}

func (mc *ManagedCluster) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.ClusterSetupType: constants.ManagedClusterSetup,
	}
}

func (mc *ManagedCluster) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// ManagedClusterList contains a list of ManagedCluster
type ManagedClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedCluster{}, &ManagedClusterList{})
}
