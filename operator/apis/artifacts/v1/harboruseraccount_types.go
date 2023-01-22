package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/harbor"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OperatorProps struct {
	HarborUser *harbor.User `json:"harborUser,omitempty"`
}

// HarborUserAccountSpec defines the desired state of HarborUserAccount
type HarborUserAccountSpec struct {
	// +kubebuilder:default=true
	Enabled          bool          `json:"enabled,omitempty"`
	ProjectRef       string        `json:"projectRef"`
	DockerConfigName string        `json:"dockerConfigName,omitempty"`
	OperatorProps    OperatorProps `json:"operatorProps,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.projectRef",name=Harbor-Project,type=string
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// HarborUserAccount is the Schema for the harboruseraccounts API
type HarborUserAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborUserAccountSpec `json:"spec,omitempty"`
	Status rApi.Status           `json:"status,omitempty"`
}

func (h *HarborUserAccount) GetStatus() *rApi.Status {
	return &h.Status
}

func (h *HarborUserAccount) GetEnsuredLabels() map[string]string {
	return map[string]string{
		"kloudlite.io/harbor-project.name": h.Spec.ProjectRef,
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
