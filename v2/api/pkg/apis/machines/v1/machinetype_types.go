package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="CPU",type=string,JSONPath=`.spec.resources.cpu`
// +kubebuilder:printcolumn:name="Memory",type=string,JSONPath=`.spec.resources.memory`
// +kubebuilder:printcolumn:name="GPU",type=string,JSONPath=`.spec.resources.gpu`
// +kubebuilder:printcolumn:name="Active",type=boolean,JSONPath=`.spec.active`
// +kubebuilder:printcolumn:name="Category",type=string,JSONPath=`.spec.category`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// MachineType represents a predefined machine configuration that users can select
type MachineType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineTypeSpec   `json:"spec,omitempty"`
	Status MachineTypeStatus `json:"status,omitempty"`
}

// MachineTypeSpec defines the desired state of MachineType
type MachineTypeSpec struct {
	// DisplayName is the human-friendly name shown to users
	// +kubebuilder:validation:Required
	DisplayName string `json:"displayName"`

	// Description provides details about this machine type
	// +optional
	Description string `json:"description,omitempty"`

	// Category groups machine types (e.g., "general", "compute-optimized", "memory-optimized", "gpu")
	// +kubebuilder:validation:Enum=general;compute-optimized;memory-optimized;gpu;development
	// +kubebuilder:default=general
	Category string `json:"category"`

	// Resources defines the compute resources for this machine type
	// +kubebuilder:validation:Required
	Resources MachineResources `json:"resources"`

	// Active determines if this machine type can be selected by users
	// +kubebuilder:default=true
	Active bool `json:"active"`

	// Priority for sorting in UI (lower numbers appear first)
	// +kubebuilder:default=100
	Priority int32 `json:"priority,omitempty"`

	// Labels to apply to WorkMachine pods
	// +optional
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// Annotations to apply to WorkMachine pods
	// +optional
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`

	// NodeSelector for pod scheduling
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations for pod scheduling
	// +optional
	Tolerations []Toleration `json:"tolerations,omitempty"`
}

// MachineResources defines the compute resources
type MachineResources struct {
	// CPU cores (e.g., "2", "4", "8")
	// +kubebuilder:validation:Required
	CPU string `json:"cpu"`

	// Memory in Gi (e.g., "4Gi", "8Gi", "16Gi")
	// +kubebuilder:validation:Required
	Memory string `json:"memory"`

	// GPU count (optional, e.g., "1", "2")
	// +optional
	GPU string `json:"gpu,omitempty"`

	// Storage for workspace (e.g., "50Gi", "100Gi")
	// +kubebuilder:validation:Required
	// +kubebuilder:default="50Gi"
	Storage string `json:"storage"`

	// EphemeralStorage for temporary files
	// +optional
	// +kubebuilder:default="10Gi"
	EphemeralStorage string `json:"ephemeralStorage,omitempty"`
}

// Toleration represents a pod toleration
type Toleration struct {
	// Key is the taint key that the toleration applies to
	// +optional
	Key string `json:"key,omitempty"`

	// Operator represents a key's relationship to the value
	// +kubebuilder:validation:Enum=Exists;Equal
	// +optional
	Operator string `json:"operator,omitempty"`

	// Value is the taint value the toleration matches to
	// +optional
	Value string `json:"value,omitempty"`

	// Effect indicates the taint effect to match
	// +kubebuilder:validation:Enum=NoSchedule;PreferNoSchedule;NoExecute
	// +optional
	Effect string `json:"effect,omitempty"`
}

// MachineTypeStatus defines the observed state of MachineType
type MachineTypeStatus struct {
	// InUseCount tracks how many WorkMachines are using this type
	// +optional
	InUseCount int32 `json:"inUseCount,omitempty"`

	// LastUpdated timestamp
	// +optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []MachineTypeCondition `json:"conditions,omitempty"`
}

// MachineTypeCondition represents a condition of the MachineType
type MachineTypeCondition struct {
	// Type of condition
	Type MachineTypeConditionType `json:"type"`

	// Status of the condition (True, False, Unknown)
	Status metav1.ConditionStatus `json:"status"`

	// LastTransitionTime is the last time the condition changed
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a unique, one-word, CamelCase reason for the condition's last transition
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable message about the last transition
	// +optional
	Message string `json:"message,omitempty"`
}

// MachineTypeConditionType represents types of MachineType conditions
type MachineTypeConditionType string

const (
	// MachineTypeConditionReady indicates the MachineType is ready for use
	MachineTypeConditionReady MachineTypeConditionType = "Ready"

	// MachineTypeConditionValidated indicates the MachineType spec is valid
	MachineTypeConditionValidated MachineTypeConditionType = "Validated"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineTypeList contains a list of MachineType
type MachineTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MachineType `json:"items"`
}