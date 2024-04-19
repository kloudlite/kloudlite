package v1

import (
	"fmt"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BlueprintSpec defines the desired state of Blueprint
type BlueprintSpec struct {
	Apps    []App  `json:"apps,omitempty"`
	Version string `json:"version"`
}

type AppStatus struct {
	Name        string `json:"name"`
	rApi.Status `json:",inline"`
}

type BlueprintStatus struct {
	rApi.Status `json:",inline"`
	Apps        []AppStatus `json:"apps,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Blueprint is the Schema for the blueprints API
type Blueprint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BlueprintSpec   `json:"spec,omitempty"`
	Status BlueprintStatus `json:"status,omitempty" graphql:"noinput"`
}

func (b *Blueprint) EnsureGVK() {
	if b != nil {
		b.SetGroupVersionKind(GroupVersion.WithKind("Blueprint"))
	}
}

func (b *Blueprint) GetStatus() *rApi.Status {
	return &b.Status.Status
}

func (b *Blueprint) GetEnsuredLabels() map[string]string {
	m := map[string]string{
		"kloudlite.io/blueprint.name": b.Name,
	}

	return m
}

func (b *Blueprint) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("Blueprint").String(),
	}
}

func (b *Blueprint) LogRef() string {
	return fmt.Sprintf("%s/%s/%s", b.Namespace, "Blueprint", b.Name)
}

//+kubebuilder:object:root=true

// BlueprintList contains a list of Blueprint
type BlueprintList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Blueprint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Blueprint{}, &BlueprintList{})
}
