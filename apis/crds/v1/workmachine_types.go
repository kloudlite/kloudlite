package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	rApi "github.com/kloudlite/operator/toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AWSMachineConfig struct {
	// Region           string `json:"region" graphql:"noinput"`
	// AvailabilityZone string `json:"availabilityZone"`

	AMI          string `json:"ami"`
	InstanceType string `json:"instanceType"`
	// PublicSubnetID string `json:"publicSubnetID"`

	//+kubebuilder:default=50
	RootVolumeSize int `json:"rootVolumeSize" graphql:"noinput"`

	//+kubebuilder:default=gp3
	RootVolumeType string `json:"rootVolumeType" graphql:"noinput"`

	ExternalVolumeSize int `json:"externalVolumeSize"`

	//+kubebuilder:default=gp3
	ExternalVolumeType string `json:"externalVolumeType" graphql:"noinput"`

	IAMInstanceProfileRole *string `json:"iamInstanceProfileRole,omitempty" graphql:"noinput"`
}

// +kubebuilder:validation:Enum=ON;OFF;
// +kubebuilder:default=ON
type WorkMachineState string

const (
	WorkMachineStateOn  WorkMachineState = "ON"
	WorkMachineStateOff WorkMachineState = "OFF"
)

type WorkMachineJobParams struct {
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
}

// WorkMachineSpec defines the desired state of WorkMachine
type WorkMachineSpec struct {
	State         WorkMachineState `json:"state"`
	SSHPublicKeys []string         `json:"sshPublicKeys"`

	JobParams WorkMachineJobParams `json:"jobParams,omitempty"`

	TargetNamespace string `json:"targetNamespace,omitempty"`

	AWSMachineConfig *AWSMachineConfig `json:"aws"`
}

func (wms *WorkMachineSpec) GetCloudProvider() ct.CloudProvider {
	if wms.AWSMachineConfig != nil {
		return ct.CloudProviderAWS
	}

	return ct.CloudProviderUnknown
}

type WorkMachineStatus struct {
	rApi.Status         `json:"status,omitempty"`
	MachinePublicSSHKey string `json:"machineSSHKey,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".spec.targetNamespace",name=TargetNamespace,type=string
// +kubebuilder:printcolumn:JSONPath=".status.status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// WorkMachine is the Schema for the workmachines API
type WorkMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkMachineSpec   `json:"spec,omitempty"`
	Status WorkMachineStatus `json:"status,omitempty"`
}

func (r *WorkMachine) EnsureGVK() {
	if r != nil {
		r.SetGroupVersionKind(GroupVersion.WithKind("WorkMachine"))
	}
}

func (w *WorkMachine) GetStatus() *rApi.Status {
	return &w.Status.Status
}

func (w *WorkMachine) GetEnsuredLabels() map[string]string {
	return map[string]string{
		"kloudlite.io/workmachine.name": w.Name,
	}
}

func (w *WorkMachine) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// WorkMachineList contains a list of WorkMachine
type WorkMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkMachine{}, &WorkMachineList{})
}
