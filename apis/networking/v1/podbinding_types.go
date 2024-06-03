package v1

import (
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodBindingSpec defines the desired state of PodBinding
type PodBindingSpec struct {
	GlobalIP     string   `json:"globalIP"`
	WgPrivateKey string   `json:"wgPrivateKey"`
	WgPublicKey  string   `json:"wgPublicKey"`
	AllowedIPs   []string `json:"allowedIPs"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:JSONPath=".spec.globalIP",name=GlobalIP,type=string
//+kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// PodBinding is the Schema for the podbindings API
type PodBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodBindingSpec `json:"spec,omitempty"`
	Status rApi.Status    `json:"status,omitempty"`
}

func (p *PodBinding) EnsureGVK() {
	if p != nil {
		p.SetGroupVersionKind(GroupVersion.WithKind("PodBinding"))
	}
}

//+kubebuilder:object:root=true

// PodBindingList contains a list of PodBinding
type PodBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodBinding{}, &PodBindingList{})
}

/*
100.160.0.0/18 => 16K IPs
100.160.0.0/19 => 8K IPs
100.160.0.0/20 => 4K IPs
100.160.0.0/21 => 2K IPs
100.160.0.0/22 => 1K IPs
100.160.0.0/23 => 512 IPs
100.160.0.0/24 => 256 IPs
*/
