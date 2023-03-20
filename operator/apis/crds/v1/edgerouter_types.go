package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EdgeRouterSpec defines the desired state of EdgeRouter
type EdgeRouterSpec struct {
	EdgeName   string `json:"edgeName"`
	AccountRef string `json:"accountRef"`
	// +kubebuilder:validation:Enum=ClusterIP;LoadBalancer
	// +kubebuilder:default=LoadBalancer
	ServiceType string `json:"serviceType,omitempty"`

	DefaultSSLCert SSLCertRef          `json:"defaultSSLCert,omitempty"`
	NodeSelector   map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations    []corev1.Toleration `json:"tolerations,omitempty"`

	WildcardDomains []string `json:"wildcardDomains,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// EdgeRouter is the Schema for the edgerouters API
type EdgeRouter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EdgeRouterSpec `json:"spec,omitempty"`
	Status rApi.Status    `json:"status,omitempty"`
}

func (er *EdgeRouter) EnsureGVK() {
	if er != nil {
		er.SetGroupVersionKind(GroupVersion.WithKind("EdgeRouter"))
	}
}

func (edge *EdgeRouter) GetStatus() *rApi.Status {
	return &edge.Status
}

func (edge *EdgeRouter) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.EdgeRouterNameKey: edge.Name,
		constants.EdgeNameKey:       edge.Spec.EdgeName,
	}
}

func (edge *EdgeRouter) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("EdgeRouter").String(),
	}
}

// +kubebuilder:object:root=true

// EdgeRouterList contains a list of EdgeRouter
type EdgeRouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EdgeRouter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EdgeRouter{}, &EdgeRouterList{})
}
