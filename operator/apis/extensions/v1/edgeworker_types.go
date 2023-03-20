package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CredentialsRef struct {
	Namespace  string `json:"namespace"`
	SecretName string `json:"secretName"`
}

type EdgeWorkerSpec struct {
	AccountName string         `json:"accountName"`
	Creds       CredentialsRef `json:"credentialsRef"`
	Provider    string         `json:"provider"`
	Region      string         `json:"region,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// EdgeWorker is the Schema for the edgeworkers API
type EdgeWorker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EdgeWorkerSpec `json:"spec,omitempty"`
	Status rApi.Status    `json:"status,omitempty"`
}

func (ew *EdgeWorker) EnsureGVK() {
	if ew != nil {
		ew.SetGroupVersionKind(GroupVersion.WithKind("EdgeWorker"))
	}
}

func (e *EdgeWorker) GetStatus() *rApi.Status {
	return &e.Status
}

func (e *EdgeWorker) GetEnsuredLabels() map[string]string {
	return map[string]string{constants.EdgeNameKey: e.Name, constants.ProviderRef: e.Spec.Creds.SecretName}
}

func (e *EdgeWorker) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("EdgeWorker").String(),
	}
}

// +kubebuilder:object:root=true

// EdgeWorkerList contains a list of EdgeWorker
type EdgeWorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EdgeWorker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EdgeWorker{}, &EdgeWorkerList{})
}
