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
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.spec.targetNamespace`
// +kubebuilder:printcolumn:name="WorkMachine",type=string,JSONPath=`.spec.workmachineName`
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

	// Name is the simple environment name (e.g., "dev-env") used for display as {userName}/{envName}
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	// +optional
	Name string `json:"name,omitempty"`

	// OwnedBy is the username of the user who owns this environment
	// +kubebuilder:validation:Required
	OwnedBy string `json:"ownedBy"`

	// Visibility controls who can see this environment
	// - private: only the owner can see
	// - shared: shared with specific users listed in SharedWith
	// - open: visible to all team members
	// +kubebuilder:validation:Enum=private;shared;open
	// +kubebuilder:default=private
	// +optional
	Visibility string `json:"visibility,omitempty"`

	// SharedWith is the list of usernames this environment is shared with
	// Only used when Visibility is "shared"
	// +optional
	SharedWith []string `json:"sharedWith,omitempty"`

	// WorkMachineName references the WorkMachine this environment belongs to
	// +kubebuilder:validation:Required
	WorkMachineName string `json:"workmachineName"`

	// Activated determines whether the environment is active (true) or inactive (false)
	// When deactivated, all deployments and statefulsets are scaled to 0
	// +kubebuilder:default=true
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

	// FromSnapshot specifies a pushed snapshot to create this environment from
	// Only snapshots with status.registryStatus.pushed=true can be used
	// This field is automatically cleared after successful restoration
	// +optional
	FromSnapshot *FromSnapshotRef `json:"fromSnapshot,omitempty"`

	// NodeName specifies the node where all environment resources should run
	// This is set from the WorkMachine's node assignment
	// +optional
	NodeName string `json:"nodeName,omitempty"`
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

	// SnapshotRestoreStatus tracks the progress of creating environment from a registry snapshot
	// +optional
	SnapshotRestoreStatus *SnapshotRestoreStatus `json:"snapshotRestoreStatus,omitempty"`

	// Hash is an 8-character hash derived from environment name and owner for DNS-safe hostnames
	// Format: hash(envName-owner)
	// +optional
	Hash string `json:"hash,omitempty"`

	// Subdomain is the subdomain assigned to this environment's workmachine (e.g., "beanbag.khost.dev")
	// +optional
	Subdomain string `json:"subdomain,omitempty"`

	// LastRestoredSnapshot tracks the last snapshot that was restored to this environment
	// Used for automatic parent lineage tracking when new snapshots are created
	// +optional
	LastRestoredSnapshot *LastRestoredSnapshotInfo `json:"lastRestoredSnapshot,omitempty"`
}

// LastRestoredSnapshotInfo tracks the last restored snapshot for lineage
type LastRestoredSnapshotInfo struct {
	// Name is the name of the snapshot that was restored
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// RestoredAt is when the snapshot was restored
	// +kubebuilder:validation:Required
	RestoredAt metav1.Time `json:"restoredAt"`
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

// FromSnapshotRef specifies a pushed snapshot to create the environment from
type FromSnapshotRef struct {
	// SnapshotName is the name of the snapshot resource to fork from
	// The snapshot must have status.registryStatus.pushed=true
	// +kubebuilder:validation:Required
	SnapshotName string `json:"snapshotName"`
}

// SnapshotRestorePhase represents the current phase of snapshot restoration
type SnapshotRestorePhase string

const (
	// SnapshotRestorePhasePending indicates restoration is pending to start
	SnapshotRestorePhasePending SnapshotRestorePhase = "Pending"

	// SnapshotRestorePhasePulling indicates snapshot is being pulled from registry
	SnapshotRestorePhasePulling SnapshotRestorePhase = "Pulling"

	// SnapshotRestorePhaseRestoring indicates K8s resources are being restored
	SnapshotRestorePhaseRestoring SnapshotRestorePhase = "Restoring"

	// SnapshotRestorePhaseDataRestoring indicates PVC data is being restored from snapshot
	SnapshotRestorePhaseDataRestoring SnapshotRestorePhase = "DataRestoring"

	// SnapshotRestorePhaseCompleted indicates restoration completed successfully
	SnapshotRestorePhaseCompleted SnapshotRestorePhase = "Completed"

	// SnapshotRestorePhaseFailed indicates restoration failed
	SnapshotRestorePhaseFailed SnapshotRestorePhase = "Failed"
)

// SnapshotRestoreStatus tracks the progress of creating environment from a registry snapshot
type SnapshotRestoreStatus struct {
	// Phase represents the current phase of snapshot restoration
	// +kubebuilder:validation:Enum=Pending;Pulling;Restoring;DataRestoring;Completed;Failed
	// +optional
	Phase SnapshotRestorePhase `json:"phase,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// SourceSnapshot is the name of the snapshot being restored from
	// +optional
	SourceSnapshot string `json:"sourceSnapshot,omitempty"`

	// ImageRef is the registry image reference being pulled
	// +optional
	ImageRef string `json:"imageRef,omitempty"`

	// SnapshotRequestName is the name of the SnapshotRequest created for pulling
	// +optional
	SnapshotRequestName string `json:"snapshotRequestName,omitempty"`

	// StartTime when restoration started
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime when restoration completed (success or failure)
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// ErrorMessage if restoration failed
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`
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

	// EnvironmentConditionForked indicates resources have been forked from source environment
	EnvironmentConditionForked EnvironmentConditionType = "Forked"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EnvironmentList contains a list of Environment
type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Environment `json:"items"`
}
