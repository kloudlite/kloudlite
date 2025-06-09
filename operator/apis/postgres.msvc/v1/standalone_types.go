package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StandaloneSpec defines the desired state of Standalone
type StandaloneSpec struct {
	ct.NodeSelectorAndTolerations `json:",inline"`
	Resources                     ct.Resources `json:"resources"`
}


//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Standalone is the Schema for the standalones API
type Standalone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StandaloneSpec   `json:"spec,omitempty"`
	Status rApi.Status      `json:"status,omitempty"`

	Output ct.ManagedServiceOutput `json:"output,omitempty"`
}

func (p *Standalone) EnsureGVK() {
  if p != nil {
    p.SetGroupVersionKind(GroupVersion.WithKind("Standalone"))
  }
}

func (p *Standalone) GetStatus() *rApi.Status {
  return &p.Status
}

func (p *Standalone) GetEnsuredLabels() map[string]string {
  return map[string]string{
    constants.MsvcNameKey: p.Name,
  }
}

func (p *Standalone) GetEnsuredAnnotations() map[string]string {
  return map[string]string{}
}

//+kubebuilder:object:root=true

// StandaloneList contains a list of Standalone
type StandaloneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Standalone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Standalone{}, &StandaloneList{})
}
