package v1

import (
	"fmt"

	ct "github.com/kloudlite/operator/apis/common-types"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodBindingSpec defines the desired state of PodBinding
type PodBindingSpec struct {
	GlobalIP     string                    `json:"globalIP"`
	WgPrivateKey string                    `json:"wgPrivateKey"`
	WgPublicKey  string                    `json:"wgPublicKey"`
	PodRef       *ct.NamespacedResourceRef `json:"podRef,omitempty"`
	PodIP        *string                   `json:"podIP,omitempty"`
	AllowedIPs   []string                  `json:"allowedIPs"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:JSONPath=".spec.globalIP",name=GlobalIP,type=string
//+kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/podbinding\\.reservation",name=Allocation,type=string
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

func (p *PodBinding) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (p *PodBinding) GetEnsuredAnnotations() map[string]string {
	key := "kloudlite.io/podbinding.reservation"
	v, ok := p.GetLabels()[key]
	if !ok || v == "false" {
		return map[string]string{key: "UnReserved"}
	}

	if p.Spec.PodRef == nil {
		return map[string]string{key: "Reserved"}
	}
	return map[string]string{key: fmt.Sprintf("Reserved (%s/%s)", p.Spec.PodRef.Namespace, p.Spec.PodRef.Name)}
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
