package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

// ProjectSpec defines the desired state of Project
type ProjectSpec struct {
	AccountName     string `json:"accountName" graphql:"noinput"`
	ClusterName     string `json:"clusterName" graphql:"noinput"`
	DisplayName     string `json:"displayName,omitempty"`
	TargetNamespace string `json:"targetNamespace"`
	Logo            string `json:"logo,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".spec.accountName",name=AccountName,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.clusterName",name=ClusterName,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.targetNamespace",name="target-namespace",type=string
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Project is the Schema for the projects API
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec `json:"spec"`
	Status rApi.Status `json:"status,omitempty" graphql:"noinput"`
}

func (p *Project) EnsureGVK() {
	if p != nil {
		p.SetGroupVersionKind(GroupVersion.WithKind("Project"))
	}
}

func (p *Project) GetStatus() *rApi.Status {
	return &p.Status
}

func (p *Project) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.ProjectNameKey:     p.Name,
		constants.AccountNameKey:     p.Spec.AccountName,
		constants.ClusterNameKey:     p.Spec.ClusterName,
		constants.TargetNamespaceKey: p.Spec.TargetNamespace,
	}
}

func (p *Project) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("Project").String(),
	}
}

// +kubebuilder:object:root=true

// ProjectList contains a list of Project
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}
