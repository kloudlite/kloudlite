package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BYOCSpec struct {
	Region             string `json:"region"`
	Provider           string `json:"provider"`
	AccountName        string `json:"accountName"`
	IncomingKafkaTopic string `json:"incomingKafkaTopic"`

	DisplayName    string   `json:"displayName,omitempty"`
	StorageClasses []string `json:"storageClasses,omitempty"`
	IngressClasses []string `json:"ingressClasses,omitempty"`
	PublicIPs      []string `json:"publicIps,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".spec.accountName",name=AccountName,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.provider",name=Provider,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.region",name=Region,type=string
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// BYOC is the Schema for the byocs API
type BYOC struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BYOCSpec    `json:"spec"`
	Status rApi.Status `json:"status,omitempty"`
}

func (b *BYOC) EnsureGVK() {
	if b != nil {
		b.SetGroupVersionKind(GroupVersion.WithKind("BYOC"))
	}
}

func (b *BYOC) GetStatus() *rApi.Status {
	return &b.Status
}

func (b *BYOC) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (b *BYOC) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: GroupVersion.WithKind("BYOC").String(),
	}
}

//+kubebuilder:object:root=true

// BYOCList contains a list of BYOC
type BYOCList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BYOC `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BYOC{}, &BYOCList{})
}
