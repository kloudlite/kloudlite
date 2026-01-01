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
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.status.snapshotType`
// +kubebuilder:printcolumn:name="Target",type=string,JSONPath=`.status.targetName`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Size",type=string,JSONPath=`.status.sizeHuman`
// +kubebuilder:printcolumn:name="Created",type=date,JSONPath=`.status.createdAt`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Snapshot represents a point-in-time snapshot of an environment's data and metadata
type Snapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotSpec   `json:"spec,omitempty"`
	Status SnapshotStatus `json:"status,omitempty"`
}

// SnapshotSpec defines the desired state of Snapshot
// Only ONE of EnvironmentRef or WorkspaceRef should be set
type SnapshotSpec struct {
	// EnvironmentRef references the environment to snapshot
	// +optional
	EnvironmentRef *EnvironmentReference `json:"environmentRef,omitempty"`

	// WorkspaceRef references the workspace to snapshot
	// +optional
	WorkspaceRef *WorkspaceReference `json:"workspaceRef,omitempty"`

	// Description is an optional description for this snapshot
	// +optional
	Description string `json:"description,omitempty"`

	// OwnedBy is the username who created this snapshot
	// +kubebuilder:validation:Required
	OwnedBy string `json:"ownedBy"`

	// IncludeMetadata controls whether to include K8s resource metadata
	// (ConfigMaps, Secrets, Deployments, etc. for environments;
	// PackageRequests and settings for workspaces)
	// +kubebuilder:default=true
	IncludeMetadata bool `json:"includeMetadata"`

	// RetentionPolicy defines when this snapshot should be deleted
	// +optional
	RetentionPolicy *RetentionPolicy `json:"retentionPolicy,omitempty"`
}

// EnvironmentReference is a reference to an environment
type EnvironmentReference struct {
	// Name is the name of the environment
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

// WorkspaceReference is a reference to a workspace
type WorkspaceReference struct {
	// Name is the name of the workspace
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// WorkmachineName is the workmachine where the workspace runs
	// +kubebuilder:validation:Required
	WorkmachineName string `json:"workmachineName"`
}

// RetentionPolicy defines when a snapshot should be automatically deleted
type RetentionPolicy struct {
	// ExpiresAt is when this snapshot should be automatically deleted
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`

	// KeepForDays is the number of days to keep the snapshot
	// +kubebuilder:validation:Minimum=1
	// +optional
	KeepForDays *int32 `json:"keepForDays,omitempty"`
}

// SnapshotState represents the current state of a snapshot
type SnapshotState string

const (
	// SnapshotStatePending means the snapshot is pending creation
	SnapshotStatePending SnapshotState = "Pending"

	// SnapshotStateCreating means the snapshot is being created
	SnapshotStateCreating SnapshotState = "Creating"

	// SnapshotStateReady means the snapshot is complete and ready
	SnapshotStateReady SnapshotState = "Ready"

	// SnapshotStateRestoring means the snapshot is being restored
	SnapshotStateRestoring SnapshotState = "Restoring"

	// SnapshotStateDeleting means the snapshot is being deleted
	SnapshotStateDeleting SnapshotState = "Deleting"

	// SnapshotStateFailed means the snapshot operation failed
	SnapshotStateFailed SnapshotState = "Failed"
)

// SnapshotType represents the type of snapshot (environment or workspace)
type SnapshotType string

const (
	// SnapshotTypeEnvironment is an environment snapshot
	SnapshotTypeEnvironment SnapshotType = "Environment"

	// SnapshotTypeWorkspace is a workspace snapshot
	SnapshotTypeWorkspace SnapshotType = "Workspace"
)

// SnapshotStatus defines the observed state of Snapshot
type SnapshotStatus struct {
	// State is the current state of the snapshot
	// +kubebuilder:default=Pending
	State SnapshotState `json:"state,omitempty"`

	// SnapshotType indicates whether this is an environment or workspace snapshot
	// +optional
	SnapshotType SnapshotType `json:"snapshotType,omitempty"`

	// TargetName is the name of the environment or workspace being snapshotted
	// Used for display in kubectl output
	// +optional
	TargetName string `json:"targetName,omitempty"`

	// Message provides human-readable status information
	// +optional
	Message string `json:"message,omitempty"`

	// SizeBytes is the total size of the snapshot in bytes
	// +optional
	SizeBytes int64 `json:"sizeBytes,omitempty"`

	// SizeHuman is the human-readable size (e.g., "1.5 GB")
	// +optional
	SizeHuman string `json:"sizeHuman,omitempty"`

	// SnapshotPath is the btrfs snapshot path on disk
	// +optional
	SnapshotPath string `json:"snapshotPath,omitempty"`

	// MetadataPath is the path to stored K8s metadata JSON
	// +optional
	MetadataPath string `json:"metadataPath,omitempty"`

	// CreatedAt is when the snapshot was successfully created
	// +optional
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// WorkMachineName is the workmachine where the snapshot is stored
	// +optional
	WorkMachineName string `json:"workMachineName,omitempty"`

	// PVCSnapshots lists individual PVC snapshot information (for environment snapshots)
	// +optional
	PVCSnapshots []PVCSnapshotInfo `json:"pvcSnapshots,omitempty"`

	// ResourceMetadata tracks captured K8s resources (for environment snapshots)
	// +optional
	ResourceMetadata *ResourceMetadataInfo `json:"resourceMetadata,omitempty"`

	// WorkspaceName is the name of the snapshotted workspace (for workspace snapshots)
	// +optional
	WorkspaceName string `json:"workspaceName,omitempty"`

	// PackageRequestsPath is the path to stored PackageRequest JSON (for workspace snapshots)
	// +optional
	PackageRequestsPath string `json:"packageRequestsPath,omitempty"`

	// WorkspaceWasSuspended indicates if the workspace was suspended during snapshot
	// +optional
	WorkspaceWasSuspended bool `json:"workspaceWasSuspended,omitempty"`

	// PreviousWorkspaceStatus stores the workspace's status before suspension
	// +optional
	PreviousWorkspaceStatus string `json:"previousWorkspaceStatus,omitempty"`
}

// PVCSnapshotInfo contains info about a snapshotted PVC
type PVCSnapshotInfo struct {
	// PVCName is the name of the PVC
	PVCName string `json:"pvcName"`

	// SnapshotPath is the path to this PVC's snapshot
	SnapshotPath string `json:"snapshotPath"`

	// SizeBytes is the size of this PVC snapshot
	SizeBytes int64 `json:"sizeBytes"`
}

// ResourceMetadataInfo tracks the count of captured K8s resources
type ResourceMetadataInfo struct {
	// ConfigMaps count
	ConfigMaps int32 `json:"configMaps"`

	// Secrets count
	Secrets int32 `json:"secrets"`

	// Deployments count
	Deployments int32 `json:"deployments"`

	// Services count
	Services int32 `json:"services"`

	// StatefulSets count
	StatefulSets int32 `json:"statefulSets"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SnapshotList contains a list of Snapshot
type SnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Snapshot `json:"items"`
}
