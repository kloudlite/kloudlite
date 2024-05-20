package v1

import (
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=CNAME;IPAddr;
type ExternalAppRecordType string

const (
	ExternalAppRecordTypeCNAME  ExternalAppRecordType = "CNAME"
	ExternalAppRecordTypeIPAddr ExternalAppRecordType = "IPAddr"
)

// ExternalAppSpec defines the desired state of ExternalApp
type ExternalAppSpec struct {
	RecordType ExternalAppRecordType `json:"recordType"`
	Record     string                `json:"record"`

	Intercept *Intercept `json:"intercept,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ExternalApp is the Schema for the externalapps API
type ExternalApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExternalAppSpec `json:"spec,omitempty"`
	Status rApi.Status     `json:"status,omitempty"`
}

func (p *ExternalApp) EnsureGVK() {
	if p != nil {
		p.SetGroupVersionKind(GroupVersion.WithKind("ExternalApp"))
	}
}

func (p *ExternalApp) GetStatus() *rApi.Status {
	return &p.Status
}

func (p *ExternalApp) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (p *ExternalApp) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// ExternalAppList contains a list of ExternalApp
type ExternalAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExternalApp{}, &ExternalAppList{})
}
