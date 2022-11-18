package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/pkg/constants"
	rApi "operators.kloudlite.io/pkg/operator"
)

// AccountRouterSpec defines the desired state of AccountRouter
type AccountRouterSpec struct {
	ControllerName string `json:"controllerName,omitempty"`
	Region         string `json:"region"`
	AccountRef     string `json:"accountRef"`

	// +kubebuilder:validation:Enum=ClusterIP;LoadBalancer
	ServiceType string `json:"serviceType"`

	DefaultSSLCert SSLCertRef        `json:"defaultSSLCert,omitempty"`
	NodeSelector   map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:default=100
	MaxBodySizeInMB int       `json:"maxBodySizeInMB,omitempty"`
	RateLimit       RateLimit `json:"rateLimit,omitempty"`
	Https           Https     `json:"https,omitempty"`
	WildcardDomains []string  `json:"wildcardDomains,omitempty"`
}

type SSLCertRef struct {
	SecretName string `json:"secretName"`
	Namespace  string `json:"namespace,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AccountRouter is the Schema for the accountrouters API
type AccountRouter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccountRouterSpec `json:"spec,omitempty"`
	Status rApi.Status       `json:"status,omitempty"`
}

func (r *AccountRouter) GetStatus() *rApi.Status {
	return &r.Status
}

func (r *AccountRouter) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.AccountRouterNameKey: r.Name,
		constants.AccountRef:           r.Spec.AccountRef,
	}
}

func (r *AccountRouter) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// AccountRouterList contains a list of AccountRouter
type AccountRouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccountRouter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AccountRouter{}, &AccountRouterList{})
}
