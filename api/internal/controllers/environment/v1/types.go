package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.spec.targetNamespace`
// +kubebuilder:printcolumn:name="Activated",type=boolean,JSONPath=`.spec.activated`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Last Activated",type=date,JSONPath=`.status.lastActivatedTime`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Environment represents a deployment environment with its own namespace
type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvironmentSpec   `json:"spec,omitempty"`
	Status EnvironmentStatus `json:"status,omitempty"`
}

// EnvironmentSpec defines the desired state of Environment
type EnvironmentSpec struct {
	// TargetNamespace is the namespace where all environment resources will be deployed
	// +kubebuilder:validation:Required
	TargetNamespace string `json:"targetNamespace"`

	// CreatedBy is the username/email of the user who created this environment
	// +kubebuilder:validation:Required
	CreatedBy string `json:"createdBy"`

	// Activated determines whether the environment is active (true) or inactive (false)
	// When deactivated, all deployments and statefulsets are scaled to 0
	// +kubebuilder:default=false
	Activated bool `json:"activated"`

	// ResourceQuotas defines resource quotas for the environment namespace
	// +optional
	ResourceQuotas *ResourceQuotas `json:"resourceQuotas,omitempty"`

	// NetworkPolicies defines network policies for the environment
	// +optional
	NetworkPolicies *NetworkPolicies `json:"networkPolicies,omitempty"`

	// Labels to apply to the namespace
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations to apply to the namespace
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// CloneFrom specifies the source environment name to clone resources from
	// This field is automatically cleared after successful cloning
	// +optional
	CloneFrom string `json:"cloneFrom,omitempty"`

	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
}

// ResourceQuotas defines resource quotas for the environment
type ResourceQuotas struct {
	// Maximum CPU limit for all pods in namespace
	// +optional
	LimitsCPU string `json:"limits.cpu,omitempty"`

	// Maximum memory limit for all pods in namespace
	// +optional
	LimitsMemory string `json:"limits.memory,omitempty"`

	// Maximum CPU requests for all pods in namespace
	// +optional
	RequestsCPU string `json:"requests.cpu,omitempty"`

	// Maximum memory requests for all pods in namespace
	// +optional
	RequestsMemory string `json:"requests.memory,omitempty"`

	// Maximum number of PVCs
	// +optional
	PersistentVolumeClaims string `json:"persistentvolumeclaims,omitempty"`

	// Maximum number of NodePort services
	// +optional
	ServicesNodePorts string `json:"services.nodeports,omitempty"`

	// Maximum number of LoadBalancer services
	// +optional
	ServicesLoadBalancers string `json:"services.loadbalancers,omitempty"`
}

// NetworkPolicies defines network policy configuration for the environment
type NetworkPolicies struct {
	// Whether to enable network policies
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`

	// List of namespaces allowed to communicate with this environment
	// +optional
	AllowedNamespaces []string `json:"allowedNamespaces,omitempty"`

	// Custom ingress rules
	// +optional
	IngressRules []IngressRule `json:"ingressRules,omitempty"`
}

// IngressRule defines a network policy ingress rule
type IngressRule struct {
	// From defines the source of the traffic
	// +optional
	From []NetworkPolicyPeer `json:"from,omitempty"`

	// Ports defines the destination ports
	// +optional
	Ports []NetworkPolicyPort `json:"ports,omitempty"`
}

// NetworkPolicyPeer defines a peer for network policy
type NetworkPolicyPeer struct {
	// NamespaceSelector selects namespaces
	// +optional
	NamespaceSelector *LabelSelector `json:"namespaceSelector,omitempty"`

	// PodSelector selects pods
	// +optional
	PodSelector *LabelSelector `json:"podSelector,omitempty"`
}

// LabelSelector is a simplified label selector
type LabelSelector struct {
	// MatchLabels is a map of labels
	// +optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

// NetworkPolicyPort defines a port for network policy
type NetworkPolicyPort struct {
	// Protocol (TCP or UDP)
	// +kubebuilder:validation:Enum=TCP;UDP
	// +optional
	Protocol string `json:"protocol,omitempty"`

	// Port number
	// +optional
	Port int32 `json:"port,omitempty"`
}

// EnvironmentStatus defines the observed state of Environment
type EnvironmentStatus struct {
	// State represents the current state of the environment
	// +kubebuilder:validation:Enum=active;inactive;activating;deactivating;deleting;error
	State EnvironmentState `json:"state,omitempty"`

	// Message provides human-readable information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// LastActivatedTime is the last time the environment was activated
	// +optional
	LastActivatedTime *metav1.Time `json:"lastActivatedTime,omitempty"`

	// LastDeactivatedTime is the last time the environment was deactivated
	// +optional
	LastDeactivatedTime *metav1.Time `json:"lastDeactivatedTime,omitempty"`

	// ResourceCount tracks the number of resources in the namespace
	// +optional
	ResourceCount *ResourceCount `json:"resourceCount,omitempty"`

	// Conditions represent the latest available observations of the environment's state
	// +optional
	Conditions []EnvironmentCondition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
}

// EnvironmentState represents the state of an environment
type EnvironmentState string

const (
	// EnvironmentStateActive means the environment is active and resources are running
	EnvironmentStateActive EnvironmentState = "active"

	// EnvironmentStateInactive means the environment is inactive and resources are scaled down
	EnvironmentStateInactive EnvironmentState = "inactive"

	// EnvironmentStateActivating means the environment is being activated
	EnvironmentStateActivating EnvironmentState = "activating"

	// EnvironmentStateDeactivating means the environment is being deactivated
	EnvironmentStateDeactivating EnvironmentState = "deactivating"

	// EnvironmentStateDeleting means the environment is being deleted
	EnvironmentStateDeleting EnvironmentState = "deleting"

	// EnvironmentStateError means there was an error with the environment
	EnvironmentStateError EnvironmentState = "error"
)

// ResourceCount tracks resource counts in the namespace
type ResourceCount struct {
	Deployments  int32 `json:"deployments,omitempty"`
	StatefulSets int32 `json:"statefulsets,omitempty"`
	Services     int32 `json:"services,omitempty"`
	ConfigMaps   int32 `json:"configmaps,omitempty"`
	Secrets      int32 `json:"secrets,omitempty"`
	PVCs         int32 `json:"pvcs,omitempty"`
}

// EnvironmentCondition represents a condition of the environment
type EnvironmentCondition struct {
	// Type of condition
	Type EnvironmentConditionType `json:"type"`

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

// EnvironmentConditionType represents types of environment conditions
type EnvironmentConditionType string

const (
	// EnvironmentConditionReady indicates the environment is ready
	EnvironmentConditionReady EnvironmentConditionType = "Ready"

	// EnvironmentConditionNamespaceCreated indicates the namespace has been created
	EnvironmentConditionNamespaceCreated EnvironmentConditionType = "NamespaceCreated"

	// EnvironmentConditionResourceQuotaApplied indicates resource quotas have been applied
	EnvironmentConditionResourceQuotaApplied EnvironmentConditionType = "ResourceQuotaApplied"

	// EnvironmentConditionNetworkPolicyApplied indicates network policies have been applied
	EnvironmentConditionNetworkPolicyApplied EnvironmentConditionType = "NetworkPolicyApplied"

	// EnvironmentConditionCloned indicates resources have been cloned from source environment
	EnvironmentConditionCloned EnvironmentConditionType = "Cloned"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EnvironmentList contains a list of Environment
type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Environment `json:"items"`
}
