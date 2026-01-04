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
)

// SnapshotRequestSpec defines the desired snapshot operation
type SnapshotRequestSpec struct {
	// Operation is the type of operation to perform
	// +kubebuilder:validation:Enum=create;delete;restore;push;pull
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

	// ParentLayers are the layer digests from parent snapshot(s) to include
	// These layers are referenced but not re-uploaded during push
	// +optional
	ParentLayers []string `json:"parentLayers,omitempty"`
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
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SnapshotRequestList contains a list of SnapshotRequest
type SnapshotRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SnapshotRequest `json:"items"`
}
