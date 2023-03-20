package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/harbor"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HarborUserAccountSpec defines the desired state of HarborUserAccount
type HarborUserAccountSpec struct {
	// +kubebuilder:default=true
	Enabled           bool   `json:"enabled,omitempty"`
	HarborProjectName string `json:"harborProjectName"`
	TargetSecret      string `json:"targetSecret"`
	DockerConfigName  string `json:"dockerConfigName,omitempty"`
	// +kubebuilder:default={push-repository,pull-repository}
	Permissions []harbor.Permission `json:"permissions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.harborProjectName",name=Harbor-Project,type=string
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// HarborUserAccount is the Schema for the harboruseraccounts API
type HarborUserAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborUserAccountSpec `json:"spec,omitempty"`
	Status rApi.Status           `json:"status,omitempty"`
}

func (h *HarborUserAccount) EnsureGVK() {
	if h != nil {
		h.SetGroupVersionKind(GroupVersion.WithKind("HarborUserAccount"))
	}
}

func (h *HarborUserAccount) GetStatus() *rApi.Status {
	return &h.Status
}

func (h *HarborUserAccount) GetEnsuredLabels() map[string]string {
	return map[string]string{
		"kloudlite.io/harbor-project.name": h.Spec.HarborProjectName,
	}
}

func (h *HarborUserAccount) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("HarborUserAccount").String(),
	}
}

// +kubebuilder:object:root=true

// HarborUserAccountList contains a list of HarborUserAccount
type HarborUserAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborUserAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HarborUserAccount{}, &HarborUserAccountList{})
}
