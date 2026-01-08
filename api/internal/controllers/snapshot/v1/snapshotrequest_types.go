package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Operation",type=string,JSONPath=`.spec.operation`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SnapshotRequest is a namespaced resource that instructs the workmachine-node-manager
// to perform btrfs snapshot operations on the local filesystem
type SnapshotRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotRequestSpec   `json:"spec,omitempty"`
	Status SnapshotRequestStatus `json:"status,omitempty"`
}

// SnapshotRequestOperation defines the type of snapshot operation
type SnapshotRequestOperation string

const (
	// SnapshotOperationCreate creates a new btrfs snapshot
	SnapshotOperationCreate SnapshotRequestOperation = "create"

	// SnapshotOperationDelete deletes an existing btrfs snapshot
	SnapshotOperationDelete SnapshotRequestOperation = "delete"

	// SnapshotOperationRestore restores data from a snapshot
	SnapshotOperationRestore SnapshotRequestOperation = "restore"

	// SnapshotOperationPush pushes a snapshot to the registry as OCI image
	SnapshotOperationPush SnapshotRequestOperation = "push"

	// SnapshotOperationPull pulls a snapshot from the registry
	SnapshotOperationPull SnapshotRequestOperation = "pull"

	// SnapshotOperationTag adds an additional tag to an existing image
	SnapshotOperationTag SnapshotRequestOperation = "tag"

	// SnapshotOperationEnsureSubvolume ensures a btrfs subvolume exists (creates if needed)
	SnapshotOperationEnsureSubvolume SnapshotRequestOperation = "ensure-subvolume"
)

// SnapshotRequestSpec defines the desired snapshot operation
type SnapshotRequestSpec struct {
	// Operation is the type of operation to perform
	// +kubebuilder:validation:Enum=create;delete;restore;push;pull;tag;ensure-subvolume
	// +kubebuilder:validation:Required
	Operation SnapshotRequestOperation `json:"operation"`

	// SourcePath is the btrfs subvolume path to snapshot (for create/restore)
	// +optional
	SourcePath string `json:"sourcePath,omitempty"`

	// SnapshotPath is where to create the snapshot or the snapshot to delete/restore from
	// +kubebuilder:validation:Required
	SnapshotPath string `json:"snapshotPath"`

	// SnapshotRef references the parent Snapshot resource
	// +optional
	SnapshotRef string `json:"snapshotRef,omitempty"`

	// EnvironmentName is the environment being snapshotted (for environment snapshots)
	// +optional
	EnvironmentName string `json:"environmentName,omitempty"`

	// WorkspaceName is the workspace being snapshotted (for workspace snapshots)
	// +optional
	WorkspaceName string `json:"workspaceName,omitempty"`

	// ReadOnly indicates whether to create a read-only snapshot (recommended)
	// +kubebuilder:default=true
	ReadOnly bool `json:"readOnly"`

	// ParentSnapshotPath is the path to the parent snapshot (for incremental push)
	// +optional
	ParentSnapshotPath string `json:"parentSnapshotPath,omitempty"`

	// RegistryRef contains registry configuration for push/pull operations
	// +optional
	RegistryRef *SnapshotRequestRegistryRef `json:"registryRef,omitempty"`

	// Metadata contains K8s resource metadata to write to the snapshot directory
	// This is used for environment snapshots to include ConfigMaps, Secrets, Deployments, etc.
	// +optional
	Metadata *SnapshotMetadata `json:"metadata,omitempty"`

	// MetadataPath is the base path where metadata should be written
	// This is separate from SnapshotPath when metadata needs to go to a parent directory
	// +optional
	MetadataPath string `json:"metadataPath,omitempty"`
}

// SnapshotMetadata contains K8s resource metadata as JSON strings
type SnapshotMetadata struct {
	// ConfigMaps JSON
	// +optional
	ConfigMaps string `json:"configMaps,omitempty"`

	// Secrets JSON
	// +optional
	Secrets string `json:"secrets,omitempty"`

	// Deployments JSON
	// +optional
	Deployments string `json:"deployments,omitempty"`

	// Services JSON
	// +optional
	Services string `json:"services,omitempty"`

	// StatefulSets JSON
	// +optional
	StatefulSets string `json:"statefulSets,omitempty"`

	// Compositions JSON
	// +optional
	Compositions string `json:"compositions,omitempty"`

	// PVCs JSON - Contains PersistentVolumeClaim specs for data restoration
	// +optional
	PVCs string `json:"pvcs,omitempty"`
}

// SnapshotRequestRegistryRef contains registry configuration for push/pull
type SnapshotRequestRegistryRef struct {
	// RegistryURL is the registry base URL (e.g., "image-registry:5000")
	// +kubebuilder:validation:Required
	RegistryURL string `json:"registryURL"`

	// Repository is the image repository path (e.g., "snapshots/username")
	// +kubebuilder:validation:Required
	Repository string `json:"repository"`

	// Tag is the image tag
	// +kubebuilder:validation:Required
	Tag string `json:"tag"`

	// ParentImageRef is the full image reference of the parent snapshot
	// Used for parent reference in v2 format (stored in config labels)
	// +optional
	ParentImageRef string `json:"parentImageRef,omitempty"`

	// SourceTag is the source tag for tag operations (copy from this tag)
	// +optional
	SourceTag string `json:"sourceTag,omitempty"`
}

// SnapshotRequestPhase represents the current phase of the request
type SnapshotRequestPhase string

const (
	// SnapshotRequestPhasePending means the request is pending
	SnapshotRequestPhasePending SnapshotRequestPhase = "Pending"

	// SnapshotRequestPhaseInProgress means the operation is in progress
	SnapshotRequestPhaseInProgress SnapshotRequestPhase = "InProgress"

	// SnapshotRequestPhaseCompleted means the operation completed successfully
	SnapshotRequestPhaseCompleted SnapshotRequestPhase = "Completed"

	// SnapshotRequestPhaseFailed means the operation failed
	SnapshotRequestPhaseFailed SnapshotRequestPhase = "Failed"
)

// SnapshotRequestStatus defines the observed state of SnapshotRequest
type SnapshotRequestStatus struct {
	// Phase is the current phase of the request
	// +kubebuilder:default=Pending
	Phase SnapshotRequestPhase `json:"phase,omitempty"`

	// Message provides human-readable status information
	// +optional
	Message string `json:"message,omitempty"`

	// SizeBytes is the size of the snapshot (for create operations)
	// +optional
	SizeBytes int64 `json:"sizeBytes,omitempty"`

	// StartedAt is when the operation started
	// +optional
	StartedAt *metav1.Time `json:"startedAt,omitempty"`

	// FinishedAt is when the operation finished
	// +optional
	FinishedAt *metav1.Time `json:"finishedAt,omitempty"`

	// Digest is the OCI image manifest digest (for push operations)
	// +optional
	Digest string `json:"digest,omitempty"`

	// LayerDigests are the digests of all layers pushed (for push operations)
	// +optional
	LayerDigests []string `json:"layerDigests,omitempty"`

	// CompressedSize is the total compressed size in bytes (for push operations)
	// +optional
	CompressedSize int64 `json:"compressedSize,omitempty"`

	// PulledMetadata contains the K8s resource metadata extracted from the OCI layer (for pull operations)
	// This is used by the environment controller to restore K8s resources without reading files from disk
	// +optional
	PulledMetadata *SnapshotMetadata `json:"pulledMetadata,omitempty"`

	// PulledSnapshots contains metadata for all snapshots pulled from registry (including parent chain)
	// This enables creating Snapshot CRs for the entire parent lineage
	// +optional
	PulledSnapshots []PulledSnapshotInfo `json:"pulledSnapshots,omitempty"`
}

// PulledSnapshotInfo contains metadata about a pulled snapshot from the registry
type PulledSnapshotInfo struct {
	// Name is the original snapshot name
	Name string `json:"name"`

	// Path is the local path where this snapshot was extracted
	Path string `json:"path"`

	// SnapshotType indicates whether this is an environment or workspace snapshot
	SnapshotType string `json:"snapshotType,omitempty"`

	// TargetName is the name of the environment or workspace that was snapshotted
	TargetName string `json:"targetName,omitempty"`

	// WorkspaceName for workspace snapshots
	WorkspaceName string `json:"workspaceName,omitempty"`

	// WorkMachineName where the original snapshot was created
	WorkMachineName string `json:"workMachineName,omitempty"`

	// OwnedBy is the owner of the original snapshot
	OwnedBy string `json:"ownedBy"`

	// ParentSnapshotName is the name of this snapshot's parent (if any)
	ParentSnapshotName string `json:"parentSnapshotName,omitempty"`

	// EnvironmentName for environment snapshots
	EnvironmentName string `json:"environmentName,omitempty"`

	// Resources contains K8s resource metadata (for environment snapshots)
	Resources *SnapshotMetadata `json:"resources,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SnapshotRequestList contains a list of SnapshotRequest
type SnapshotRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SnapshotRequest `json:"items"`
}
