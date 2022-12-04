package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rApi "operators.kloudlite.io/pkg/operator"
)

type NamespacedReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// PrimaryClusterSpec defines the desired state of PrimaryCluster
type PrimaryClusterSpec struct {
	ImgPullSecrets    []NamespacedReference `json:"imagePullSecrets"`
	LokiValues        LokiValues            `json:"loki"`
	PrometheusValues  PrometheusValues      `json:"prometheus"`
	CertManagerValues CertManagerValues     `json:"certManager,omitempty"`
	IngressValues     IngressValues         `json:"ingress"`
	Operators         Operators             `json:"operators"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// PrimaryCluster is the Schema for the primaryclusters API
type PrimaryCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrimaryClusterSpec `json:"spec,omitempty"`
	Status rApi.Status        `json:"status,omitempty"`
}

func (p *PrimaryCluster) GetStatus() *rApi.Status {
	return &p.Status
}

func (p *PrimaryCluster) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (p *PrimaryCluster) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// PrimaryClusterList contains a list of PrimaryCluster
type PrimaryClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrimaryCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PrimaryCluster{}, &PrimaryClusterList{})
}
