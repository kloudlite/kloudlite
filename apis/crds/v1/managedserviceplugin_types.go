package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManagedServicePluginSpec defines the desired state of ManagedServicePlugin
type ManagedServicePluginSpec struct {
	APIVersion string                     `json:"apiVersion"`
	Kinds      []ManagedServicePluginKind `json:"kinds,omitempty"`
}

type ManagedServicePluginKind struct {
	Kind    string            `json:"kind"`
	Inputs  []MsvcPluginInput `json:"inputs"`
	Outputs MsvcPluginOutput  `json:"outputs"`
}

type MsvcPluginInputType string

const (
	InputTypeString       MsvcPluginInputType = "string"
	InputTypeInteger      MsvcPluginInputType = "integer"
	InputTypeIntegerRange MsvcPluginInputType = "number-range"
	InputTypeFloat        MsvcPluginInputType = "float"
	InputTypeFloatRange   MsvcPluginInputType = "float-range"
	InputTypeMap          MsvcPluginInputType = "map"
	InputTypeArray        MsvcPluginInputType = "array"
)

type MsvcPluginInput struct {
	Input       string              `json:"input"`
	Label       string              `json:"label,omitempty"`
	Description string              `json:"description,omitempty"`
	Type        MsvcPluginInputType `json:"type"`
	Required    bool                `json:"required"`
}

type MsvcOutputKey struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type MsvcPluginOutput struct {
	Keys []MsvcOutputKey `json:"keys"`
}

// ManagedServicePluginStatus defines the observed state of ManagedServicePlugin
type ManagedServicePluginStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// ManagedServicePlugin is the Schema for the managedserviceplugins API
type ManagedServicePlugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedServicePluginSpec   `json:"spec,omitempty"`
	Status ManagedServicePluginStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ManagedServicePluginList contains a list of ManagedServicePlugin
type ManagedServicePluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedServicePlugin `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedServicePlugin{}, &ManagedServicePluginList{})
}
