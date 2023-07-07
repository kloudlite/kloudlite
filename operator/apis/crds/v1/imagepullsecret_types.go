package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImagePullSecretSpec defines the desired state of ImagePullSecret
type ImagePullSecretSpec struct {
	DockerConfigJson *string `json:"dockerConfigJson,omitempty"`

	DockerUsername         *string `json:"dockerUsername,omitempty"`
	DockerPassword         *string `json:"dockerPassword,omitempty"`
	DockerRegistryEndpoint *string `json:"dockerRegistryEndpoint,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ImagePullSecret is the Schema for the imagepullsecrets API
type ImagePullSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImagePullSecretSpec `json:"spec"`
	Status rApi.Status         `json:"status,omitempty"`
}

func (r *ImagePullSecret) EnsureGVK() {
	if r != nil {
		r.SetGroupVersionKind(GroupVersion.WithKind("ImagePullSecret"))
	}
}

func (r *ImagePullSecret) GetStatus() *rApi.Status {
	return &r.Status
}

func (r *ImagePullSecret) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.ImagePullSecretNameKey: r.Name,
	}
}

func (m *ImagePullSecret) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// ImagePullSecretList contains a list of ImagePullSecret
type ImagePullSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImagePullSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImagePullSecret{}, &ImagePullSecretList{})
}
