package v1

import (
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkmachineJobParams struct {
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
}

type WorkmachineType string

const (
	// NoOpWorkmachine does not perform any workmachine related Operation
	NoOpWorkmachine WorkmachineType = "kloudlite/no-op"

	// K3sWorkmachine creates VMs based on your Cloud Provider settings, installs K3s on it, and makes that VM part of your k3s cluster
	K3sWorkmachine WorkmachineType = "kloudlite/k3s"

	// EKSWorkmachine creates a VM/EC2 instance for your Amazon AWS EKS Cluster
	EKSWorkmachine WorkmachineType = "aws/eks"

	// GKEWorkmachine creates a VM/Compute instance for your Google Cloud GKE Cluster
	GKEWorkmachine WorkmachineType = "gcp/gke"
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

type WorkmachineSSH struct {
	Secret     corev1.LocalObjectReference `json:"secret,omitempty"`
	PublicKeys []string                    `json:"publicKeys,omitempty"`
}

type WorkmachineK3sParams struct {
	Version    string `json:"version"`
	ServerHost string `json:"serverHost"`
	AgentToken string `json:"agentToken"`
}

// WorkmachineSpec defines the desired state of Workmachine.
type WorkmachineSpec struct {
	Paused bool           `json:"paused,omitempty"`
	SSH    WorkmachineSSH `json:"ssh,omitempty"`

	JobParams WorkmachineJobParams `json:"jobParams,omitempty"`

	TargetNamespace string `json:"targetNamespace,omitempty"`

	Type WorkmachineType `json:"type"`

	AWSMachineConfig *AWSMachineConfig `json:"aws,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Workmachine is the Schema for the workmachines API.
type Workmachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkmachineSpec   `json:"spec,omitempty"`
	Status reconciler.Status `json:"status,omitempty"`
}

func (r *Workmachine) EnsureGVK() {
	if r != nil {
		r.SetGroupVersionKind(GroupVersion.WithKind("Workmachine"))
	}
}

func (w *Workmachine) GetStatus() *reconciler.Status {
	return &w.Status
}

func (w *Workmachine) GetEnsuredLabels() map[string]string {
	return map[string]string{WorkspaceNameKey: w.Name}
}

func (w *Workmachine) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// WorkmachineList contains a list of Workmachine.
type WorkmachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workmachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workmachine{}, &WorkmachineList{})
}
