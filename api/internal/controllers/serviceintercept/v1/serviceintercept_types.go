package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PortMapping defines mapping between service and workspace ports
type PortMapping struct {
	// ServicePort is the port exposed by the service
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	ServicePort int32 `json:"servicePort"`

	// WorkspacePort is the port in the workspace pod
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	WorkspacePort int32 `json:"workspacePort"`

	// Protocol is the protocol used (TCP/UDP)
	// +kubebuilder:validation:Enum=TCP;UDP;SCTP
	// +kubebuilder:default=TCP
	// +optional
	Protocol corev1.Protocol `json:"protocol,omitempty"`
}

// ServiceInterceptSpec defines the desired state of ServiceIntercept
type ServiceInterceptSpec struct {
	// WorkspaceRef references the workspace pod that will intercept traffic
	// +kubebuilder:validation:Required
	WorkspaceRef corev1.ObjectReference `json:"workspaceRef"`

	// ServiceRef references the service to intercept
	// The intercept will run in the service's namespace
	// +kubebuilder:validation:Required
	ServiceRef corev1.ObjectReference `json:"serviceRef"`

	// PortMappings defines how service ports map to workspace ports
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	PortMappings []PortMapping `json:"portMappings"`
}

// ServiceInterceptStatus defines the observed state of ServiceIntercept
type ServiceInterceptStatus struct {
	// Phase represents the current phase of the service intercept
	// +kubebuilder:validation:Enum=Creating;Active;Failed
	Phase string `json:"phase,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// OriginalServiceSelector stores the original service selector to identify pods to delete
	// +optional
	OriginalServiceSelector map[string]string `json:"originalServiceSelector,omitempty"`

	// AffectedPodNames lists pods that have been deleted/affected by the intercept
	// +optional
	AffectedPodNames []string `json:"affectedPodNames,omitempty"`

	// WorkspacePodIP is the IP address of the workspace pod
	// +optional
	WorkspacePodIP string `json:"workspacePodIP,omitempty"`

	// WorkspacePodName is the name of the workspace pod
	// +optional
	WorkspacePodName string `json:"workspacePodName,omitempty"`

	// SOCATPodName is the name of the SOCAT forwarding pod
	// +optional
	SOCATPodName string `json:"socatPodName,omitempty"`

	// WorkspaceHeadlessServiceName is the name of the headless service for the workspace
	// +optional
	WorkspaceHeadlessServiceName string `json:"workspaceHeadlessServiceName,omitempty"`

	// InterceptStartTime when the intercept was activated
	// +optional
	InterceptStartTime *metav1.Time `json:"interceptStartTime,omitempty"`

	// Conditions represent the latest available observations of service intercept state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={kloudlite,intercepts}
// +kubebuilder:printcolumn:name="Workspace",type=string,JSONPath=`.spec.workspaceRef.name`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.spec.serviceRef.name`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ServiceIntercept is the Schema for the serviceintercepts API
type ServiceIntercept struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceInterceptSpec   `json:"spec,omitempty"`
	Status ServiceInterceptStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceInterceptList contains a list of ServiceIntercept
type ServiceInterceptList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceIntercept `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceIntercept{}, &ServiceInterceptList{})
}
