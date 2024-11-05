package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ACLUserSpec defines the desired state of ACLUser
type ACLUserSpec struct {
	AdminSecretRef ct.SecretRef `json:"adminSecretRef"`
	Topics         []string     `json:"topics"`
	ResourceName   string       `json:"resourceName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ACLUser is the Schema for the aclusers API
type ACLUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ACLUserSpec `json:"spec"`
	Status rApi.Status `json:"status,omitempty"`
}

func (user *ACLUser) EnsureGVK() {
	if user != nil {
		user.SetGroupVersionKind(GroupVersion.WithKind("ACLUser"))
	}
}

func (user *ACLUser) GetStatus() *rApi.Status {
	return &user.Status
}

func (user *ACLUser) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.MsvcNameKey: "redpanda",
	}
}

func (user *ACLUser) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("ACLUser").String(),
	}
}

// +kubebuilder:object:root=true

// ACLUserList contains a list of ACLUser
type ACLUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ACLUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ACLUser{}, &ACLUserList{})
}
