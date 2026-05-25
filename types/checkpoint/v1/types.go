package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ============================================================================
// Checkpoint - Point-in-time snapshot of a btrfs subvolume
// ============================================================================

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.spec.owner`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Size",type=string,JSONPath=`.status.sizeHuman`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Checkpoint represents a point-in-time snapshot of a btrfs subvolume.
// When uploaded to an OCI registry, the data is stored as a compressed layer.
// Checkpoints are namespaced and can be owned by workspaces or environments.
type Checkpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CheckpointSpec   `json:"spec,omitempty"`
	Status CheckpointStatus `json:"status,omitempty"`
}

// CheckpointSpec defines the checkpoint parameters
type CheckpointSpec struct {
	// Owner identifies who owns this checkpoint (e.g., username)
	// +kubebuilder:validation:Required
	Owner string `json:"owner"`

	// Description is a human-readable description
	// +optional
	Description string `json:"description,omitempty"`

	// SourcePath is the btrfs subvolume path to snapshot
	// +kubebuilder:validation:Required
	SourcePath string `json:"sourcePath"`

	// NodeName is the Kubernetes node where the btrfs subvolume exists
	// +kubebuilder:validation:Required
	NodeName string `json:"nodeName"`

	// RetentionPolicy defines automatic deletion rules
	// +optional
	RetentionPolicy *RetentionPolicy `json:"retentionPolicy,omitempty"`

	// Artifacts define metadata stored alongside the checkpoint
	// +optional
	Artifacts []ArtifactSpec `json:"artifacts,omitempty"`
}

// RetentionPolicy defines when a checkpoint should be automatically deleted
type RetentionPolicy struct {
	// ExpiresAt is when this checkpoint should be automatically deleted
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`

	// KeepForDays is the number of days to keep the checkpoint
	// +kubebuilder:validation:Minimum=1
	// +optional
	KeepForDays *int32 `json:"keepForDays,omitempty"`
}

// ArtifactSpec defines a metadata artifact stored with the checkpoint
type ArtifactSpec struct {
	// Name identifies this artifact (e.g., "k8s-resources", "app-config")
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Type hints at how to handle this artifact during restore
	// +kubebuilder:validation:Enum=kubernetes-manifests;json;yaml;sql;raw
	// +kubebuilder:default=raw
	Type ArtifactType `json:"type,omitempty"`

	// Data is the artifact content (base64 encoded for binary safety)
	// +optional
	Data string `json:"data,omitempty"`
}

// ArtifactType represents the type of artifact
type ArtifactType string

const (
	ArtifactTypeKubernetesManifests ArtifactType = "kubernetes-manifests"
	ArtifactTypeJSON                ArtifactType = "json"
	ArtifactTypeYAML                ArtifactType = "yaml"
	ArtifactTypeSQL                 ArtifactType = "sql"
	ArtifactTypeRaw                 ArtifactType = "raw"
)

// CheckpointPhase represents the current phase of a checkpoint
type CheckpointPhase string

const (
	CheckpointPhasePending   CheckpointPhase = "Pending"
	CheckpointPhaseCreating  CheckpointPhase = "Creating"
	CheckpointPhaseUploading CheckpointPhase = "Uploading"
	CheckpointPhaseReady     CheckpointPhase = "Ready"
	CheckpointPhaseDeleting  CheckpointPhase = "Deleting"
	CheckpointPhaseFailed    CheckpointPhase = "Failed"
)

// CheckpointStatus defines the observed state of Checkpoint
type CheckpointStatus struct {
	// Phase is the current phase of the checkpoint
	// +kubebuilder:default=Pending
	Phase CheckpointPhase `json:"phase,omitempty"`

	// Message provides human-readable status information
	// +optional
	Message string `json:"message,omitempty"`

	// StartedAt is when processing started
	// +optional
	StartedAt *metav1.Time `json:"startedAt,omitempty"`

	// CompletedAt is when processing completed (success or failure)
	// +optional
	CompletedAt *metav1.Time `json:"completedAt,omitempty"`

	// SizeBytes is the uncompressed size in bytes
	// +optional
	SizeBytes int64 `json:"sizeBytes,omitempty"`

	// SizeHuman is the human-readable size (e.g., "1.5 GB")
	// +optional
	SizeHuman string `json:"sizeHuman,omitempty"`

	// Layer contains OCI registry storage information for the uploaded data
	// +optional
	Layer *LayerInfo `json:"layer,omitempty"`

	// LocalPath is the cached btrfs snapshot path on disk
	// +optional
	LocalPath string `json:"localPath,omitempty"`

	// ContentHash is the MD5 hash of the snapshot content (used for deduplication and as OCI tag)
	// +optional
	ContentHash string `json:"contentHash,omitempty"`

	// Artifacts lists the stored artifacts and their status
	// +optional
	Artifacts []ArtifactStatus `json:"artifacts,omitempty"`
}

// LayerInfo contains OCI registry storage details for the checkpoint data
type LayerInfo struct {
	// ImageRef is the full image reference (registry/repo:tag)
	ImageRef string `json:"imageRef"`

	// Digest is the image manifest digest (sha256:...)
	// +optional
	Digest string `json:"digest,omitempty"`

	// PushedAt is when the layer was pushed to registry
	// +optional
	PushedAt *metav1.Time `json:"pushedAt,omitempty"`

	// CompressedSize is the compressed size in bytes
	// +optional
	CompressedSize int64 `json:"compressedSize,omitempty"`
}

// ArtifactStatus tracks a stored artifact
type ArtifactStatus struct {
	Name      string       `json:"name"`
	Type      ArtifactType `json:"type"`
	SizeBytes int64        `json:"sizeBytes,omitempty"`
	Stored    bool         `json:"stored"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CheckpointList contains a list of Checkpoint
type CheckpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Checkpoint `json:"items"`
}

// ============================================================================
// CheckpointRestore - Request to restore a checkpoint
// ============================================================================

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Checkpoint",type=string,JSONPath=`.spec.checkpointName`
// +kubebuilder:printcolumn:name="Target",type=string,JSONPath=`.spec.targetPath`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// CheckpointRestore requests restoration of a checkpoint to a target path
type CheckpointRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CheckpointRestoreSpec   `json:"spec,omitempty"`
	Status CheckpointRestoreStatus `json:"status,omitempty"`
}

// CheckpointRestoreSpec defines the restore operation
type CheckpointRestoreSpec struct {
	// CheckpointName is the checkpoint to restore
	// +kubebuilder:validation:Required
	CheckpointName string `json:"checkpointName"`

	// TargetPath is where to restore the checkpoint
	// +kubebuilder:validation:Required
	TargetPath string `json:"targetPath"`

	// NodeName is the node where to perform the restore
	// +kubebuilder:validation:Required
	NodeName string `json:"nodeName"`

	// IncludeArtifacts lists which artifacts to include in response (empty = all)
	// +optional
	IncludeArtifacts []string `json:"includeArtifacts,omitempty"`
}

// CheckpointRestorePhase represents restore operation phase
type CheckpointRestorePhase string

const (
	CheckpointRestorePhasePending     CheckpointRestorePhase = "Pending"
	CheckpointRestorePhaseDownloading CheckpointRestorePhase = "Downloading"
	CheckpointRestorePhaseRestoring   CheckpointRestorePhase = "Restoring"
	CheckpointRestorePhaseCompleted   CheckpointRestorePhase = "Completed"
	CheckpointRestorePhaseFailed      CheckpointRestorePhase = "Failed"
)

// CheckpointRestoreStatus defines restore progress
type CheckpointRestoreStatus struct {
	// Phase is the current restore phase
	// +kubebuilder:default=Pending
	Phase CheckpointRestorePhase `json:"phase,omitempty"`

	// Message provides status details
	// +optional
	Message string `json:"message,omitempty"`

	// StartedAt is when the restore started
	// +optional
	StartedAt *metav1.Time `json:"startedAt,omitempty"`

	// CompletedAt is when the restore completed
	// +optional
	CompletedAt *metav1.Time `json:"completedAt,omitempty"`

	// RestoredPath is the actual path where data was restored
	// +optional
	RestoredPath string `json:"restoredPath,omitempty"`

	// Artifacts contains the retrieved artifact data
	// +optional
	Artifacts map[string]string `json:"artifacts,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CheckpointRestoreList contains a list of CheckpointRestore
type CheckpointRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CheckpointRestore `json:"items"`
}
