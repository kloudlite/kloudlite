package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Port struct {
	Port       int32 `json:"port,omitempty"`
	TargetPort int32 `json:"targetPort,omitempty"`
}

type CNameRecord struct {
	Host   string `json:"host,omitempty"`
	Target string `json:"target,omitempty"`
}

// DeviceSpec defines the desired state of Device
type DeviceSpec struct {
	Ports           []Port        `json:"ports,omitempty"`
	ActiveNamespace *string       `json:"activeNamespace,omitempty"`
	CNameRecords    []CNameRecord `json:"cnameRecords,omitempty"`

	NodeSelector      map[string]string `json:"nodeSelector,omitempty" graphql:"noinput"`
	Disabled          bool              `json:"disabled,omitempty" graphql:"noinput"`
	NoExternalService bool              `json:"noExternalService,omitempty" graphql:"noinput"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/active\\.namespace",name=Active_Ns,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Device is the Schema for the devices API
type Device struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceSpec  `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty" graphql:"noinput"`
}

func (d *Device) EnsureGVK() {
	if d != nil {
		d.SetGroupVersionKind(GroupVersion.WithKind("Device"))
	}
}

func (d *Device) GetStatus() *rApi.Status {
	return &d.Status
}

func (d *Device) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.WGDeviceNameKey: d.Name,
	}
}

func (d *Device) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: GroupVersion.WithKind("Device").String(),
		"kloudlite.io/active.namespace": func() string {
			if d.Spec.ActiveNamespace == nil {
				return ""
			}
			return *d.Spec.ActiveNamespace
		}(),
		"kloudlite.io/enabled": func() string {
			if d.Spec.Disabled {
				return "false"
			}
			return "false"
		}(),
	}
}

//+kubebuilder:object:root=true

// DeviceList contains a list of Device
type DeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Device `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Device{}, &DeviceList{})
}
