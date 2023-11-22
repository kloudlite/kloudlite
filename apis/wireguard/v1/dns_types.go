package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DnsSpec defines the desired state of Dns
type DnsSpec struct {
	MainDns *string `json:"mainDns,omitempty"`
	DNS     *string `json:"dns,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Dns is the Schema for the dns API
type Dns struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DnsSpec     `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty" graphql:"noinput"`
}

func (d *Dns) EnsureGVK() {
	if d != nil {
		d.SetGroupVersionKind(GroupVersion.WithKind("Dns"))
	}
}

func (d *Dns) GetStatus() *rApi.Status {
	return &d.Status
}

func (d *Dns) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (d *Dns) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: GroupVersion.WithKind("Dns").String(),
	}
}

//+kubebuilder:object:root=true

// DnsList contains a list of Dns
type DnsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dns `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dns{}, &DnsList{})
}
