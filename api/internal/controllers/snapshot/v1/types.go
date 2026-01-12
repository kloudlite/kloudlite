package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ============================================================================
// SnapshotStore - Storage backend configuration (OCI Registry)
// ============================================================================

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Registry",type=string,JSONPath=`.spec.registry.endpoint`
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SnapshotStore defines an OCI registry for storing snapshots
type SnapshotStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotStoreSpec   `json:"spec,omitempty"`
	Status SnapshotStoreStatus `json:"status,omitempty"`
}

// SnapshotStoreSpec defines the OCI registry configuration
type SnapshotStoreSpec struct {
	// Registry configures the OCI registry endpoint
	// +kubebuilder:validation:Required
	Registry RegistryConfig `json:"registry"`
}

// RegistryConfig configures OCI registry connection
type RegistryConfig struct {
	// Endpoint is the registry URL (e.g., "image-registry.kloudlite.svc.cluster.local:5000")
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// Insecure allows HTTP connections (for internal registries)
	// +kubebuilder:default=true
	Insecure bool `json:"insecure,omitempty"`

	// RepositoryPrefix is prepended to all snapshot repositories
	// e.g., "snapshots" results in "snapshots/{owner}/{name}"
	// +kubebuilder:default=snapshots
	RepositoryPrefix string `json:"repositoryPrefix,omitempty"`
}

// SnapshotStoreStatus defines the observed state
type SnapshotStoreStatus struct {
	// Ready indicates if the registry is accessible
	Ready bool `json:"ready,omitempty"`

	// Message provides status details
	// +optional
	Message string `json:"message,omitempty"`

	// LastChecked is when connectivity was last verified
	// +optional
	LastChecked *metav1.Time `json:"lastChecked,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SnapshotStoreList contains a list of SnapshotStore
type SnapshotStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnapshotStore `json:"items"`
}

// ============================================================================
// Snapshot - Global metadata about a stored snapshot (result of SnapshotRequest)
// ============================================================================

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.spec.owner`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="References",type=string,JSONPath=`.status.referencedBy`
// +kubebuilder:printcolumn:name="Size",type=string,JSONPath=`.status.sizeHuman`
// +kubebuilder:printcolumn:name="Parent",type=string,JSONPath=`.spec.parentSnapshot`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Snapshot represents metadata about a snapshot stored in an OCI registry.
// This is purely metadata - it doesn't know about nodes or local paths.
// Snapshots are created by SnapshotRequest after data is pushed to registry.
type Snapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotSpec   `json:"spec,omitempty"`
	Status SnapshotStatus `json:"status,omitempty"`
}

// SnapshotSpec defines the snapshot metadata
type SnapshotSpec struct {
	// Owner identifies who owns this snapshot (e.g., username)
	// +kubebuilder:validation:Required
	Owner string `json:"owner"`

	// ParentSnapshot is the parent snapshot name for incremental storage
	// +optional
	ParentSnapshot string `json:"parentSnapshot,omitempty"`

	// Description is a human-readable description
	// +optional
	Description string `json:"description,omitempty"`

	// Artifacts define metadata stored alongside the snapshot
	// +optional
	Artifacts []ArtifactSpec `json:"artifacts,omitempty"`

	// RetentionPolicy defines automatic deletion rules
	// +optional
	RetentionPolicy *RetentionPolicy `json:"retentionPolicy,omitempty"`
}

// ArtifactSpec defines a metadata artifact stored with the snapshot
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
	SnapshotStateReady    SnapshotState = "Ready"
	SnapshotStateDeleting SnapshotState = "Deleting"
	SnapshotStateFailed   SnapshotState = "Failed"
)

// SnapshotStatus defines the observed state of Snapshot
type SnapshotStatus struct {
	// State is the current state of the snapshot
	// +kubebuilder:default=Ready
	State SnapshotState `json:"state,omitempty"`

	// Message provides human-readable status information
	// +optional
	Message string `json:"message,omitempty"`

	// ReferencedBy lists the environment/workspace names that reference this snapshot
	// Snapshot cannot be garbage collected when this list is non-empty
	// +optional
	ReferencedBy []string `json:"referencedBy,omitempty"`

	// SizeBytes is the snapshot size in bytes
	// +optional
	SizeBytes int64 `json:"sizeBytes,omitempty"`

	// SizeHuman is the human-readable size (e.g., "1.5 GB")
	// +optional
	SizeHuman string `json:"sizeHuman,omitempty"`

	// CreatedAt is when the snapshot was successfully created
	// +optional
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// Lineage is the ordered list of parent snapshots (root first)
	// +optional
	Lineage []string `json:"lineage,omitempty"`

	// StorageRefs lists all registry imageRefs this snapshot needs for restore
	// Includes own imageRef plus all parent imageRefs in the chain
	// Used for garbage collection - storage is only deleted when no snapshot references it
	// +optional
	StorageRefs []string `json:"storageRefs,omitempty"`

	// Registry contains OCI registry storage information
	// +optional
	Registry *SnapshotRegistryInfo `json:"registry,omitempty"`

	// Artifacts lists the stored artifacts
	// +optional
	Artifacts []ArtifactStatus `json:"artifacts,omitempty"`
}

// SnapshotRegistryInfo contains OCI registry storage details
type SnapshotRegistryInfo struct {
	// ImageRef is the full image reference (registry/repo:tag)
	ImageRef string `json:"imageRef"`

	// Digest is the image manifest digest (sha256:...)
	// +optional
	Digest string `json:"digest,omitempty"`

	// PushedAt is when the snapshot was pushed to registry
	// +optional
	PushedAt *metav1.Time `json:"pushedAt,omitempty"`

	// CompressedSize is the total compressed size in bytes
	// +optional
	CompressedSize int64 `json:"compressedSize,omitempty"`

	// ParentImageRef is the parent snapshot's image reference
	// Used for incremental restore
	// +optional
	ParentImageRef string `json:"parentImageRef,omitempty"`
}

// ArtifactStatus tracks a stored artifact
type ArtifactStatus struct {
	Name      string       `json:"name"`
	Type      ArtifactType `json:"type"`
	SizeBytes int64        `json:"sizeBytes,omitempty"`
	Stored    bool         `json:"stored"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SnapshotList contains a list of Snapshot
type SnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Snapshot `json:"items"`
}

// ============================================================================
// SnapshotRequest - Node-specific request to create a snapshot
// ============================================================================

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Snapshot",type=string,JSONPath=`.spec.snapshotName`
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.spec.nodeName`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SnapshotRequest is a node-specific request to create a snapshot.
// The node-manager watches these and creates the actual Snapshot after
// pushing data to the registry.
type SnapshotRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotRequestSpec   `json:"spec,omitempty"`
	Status SnapshotRequestStatus `json:"status,omitempty"`
}

// SnapshotRequestSpec defines the snapshot creation request
type SnapshotRequestSpec struct {
	// SnapshotName is the name of the Snapshot to create
	// +kubebuilder:validation:Required
	SnapshotName string `json:"snapshotName"`

	// SourcePath is the btrfs subvolume path to snapshot
	// +kubebuilder:validation:Required
	SourcePath string `json:"sourcePath"`

	// NodeName is the Kubernetes node where the btrfs subvolume exists
	// +kubebuilder:validation:Required
	NodeName string `json:"nodeName"`

	// Store is the name of the SnapshotStore to use
	// +kubebuilder:validation:Required
	Store string `json:"store"`

	// Owner identifies who owns this snapshot (e.g., username)
	// +kubebuilder:validation:Required
	Owner string `json:"owner"`

	// ParentSnapshot is the parent snapshot name for incremental send
	// +optional
	ParentSnapshot string `json:"parentSnapshot,omitempty"`

	// Description is a human-readable description
	// +optional
	Description string `json:"description,omitempty"`

	// Artifacts define metadata to store alongside the snapshot
	// +optional
	Artifacts []ArtifactSpec `json:"artifacts,omitempty"`

	// RetentionPolicy defines automatic deletion rules for the created snapshot
	// +optional
	RetentionPolicy *RetentionPolicy `json:"retentionPolicy,omitempty"`
}

// SnapshotRequestState represents the current state of a snapshot request
type SnapshotRequestState string

const (
	SnapshotRequestStatePending   SnapshotRequestState = "Pending"
	SnapshotRequestStateCreating  SnapshotRequestState = "Creating"
	SnapshotRequestStateUploading SnapshotRequestState = "Uploading"
	SnapshotRequestStateCompleted SnapshotRequestState = "Completed"
	SnapshotRequestStateFailed    SnapshotRequestState = "Failed"
)

// SnapshotRequestStatus defines the observed state of SnapshotRequest
type SnapshotRequestStatus struct {
	// State is the current state of the request
	// +kubebuilder:default=Pending
	State SnapshotRequestState `json:"state,omitempty"`

	// Message provides human-readable status information
	// +optional
	Message string `json:"message,omitempty"`

	// StartedAt is when processing started
	// +optional
	StartedAt *metav1.Time `json:"startedAt,omitempty"`

	// CompletedAt is when processing completed (success or failure)
	// +optional
	CompletedAt *metav1.Time `json:"completedAt,omitempty"`

	// CreatedSnapshot is the name of the Snapshot that was created
	// +optional
	CreatedSnapshot string `json:"createdSnapshot,omitempty"`

	// LocalSnapshotPath is the temporary local btrfs snapshot path
	// +optional
	LocalSnapshotPath string `json:"localSnapshotPath,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SnapshotRequestList contains a list of SnapshotRequest
type SnapshotRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnapshotRequest `json:"items"`
}

// ============================================================================
// SnapshotRef - Reference counting mechanism
// ============================================================================

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Snapshot",type=string,JSONPath=`.spec.snapshotName`
// +kubebuilder:printcolumn:name="Bound",type=boolean,JSONPath=`.status.bound`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SnapshotRef creates a reference to a Snapshot, incrementing its refCount
// When deleted (manually or via ownerReference), decrements the refCount
// Use ownerReferences to automatically delete when the owning resource is deleted
type SnapshotRef struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotRefSpec   `json:"spec,omitempty"`
	Status SnapshotRefStatus `json:"status,omitempty"`
}

// SnapshotRefSpec defines which snapshot this ref points to
type SnapshotRefSpec struct {
	// SnapshotName is the name of the Snapshot to reference
	// +kubebuilder:validation:Required
	SnapshotName string `json:"snapshotName"`

	// Purpose describes why this reference exists (for documentation)
	// +optional
	Purpose string `json:"purpose,omitempty"`
}

// SnapshotRefStatus tracks the reference status
type SnapshotRefStatus struct {
	// Bound indicates if the refCount has been incremented
	Bound bool `json:"bound,omitempty"`

	// BoundAt is when the ref was bound to the snapshot
	// +optional
	BoundAt *metav1.Time `json:"boundAt,omitempty"`

	// SnapshotState is the current state of the referenced snapshot
	// +optional
	SnapshotState SnapshotState `json:"snapshotState,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SnapshotRefList contains a list of SnapshotRef
type SnapshotRefList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnapshotRef `json:"items"`
}

// ============================================================================
// SnapshotRestore - Request to restore a snapshot
// ============================================================================

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Snapshot",type=string,JSONPath=`.spec.snapshotName`
// +kubebuilder:printcolumn:name="Target",type=string,JSONPath=`.spec.targetPath`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SnapshotRestore requests restoration of a snapshot to a target path
type SnapshotRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotRestoreSpec   `json:"spec,omitempty"`
	Status SnapshotRestoreStatus `json:"status,omitempty"`
}

// SnapshotRestoreSpec defines the restore operation
type SnapshotRestoreSpec struct {
	// SnapshotName is the snapshot to restore
	// +kubebuilder:validation:Required
	SnapshotName string `json:"snapshotName"`

	// TargetPath is where to restore the snapshot
	// +kubebuilder:validation:Required
	TargetPath string `json:"targetPath"`

	// NodeName is the node where to perform the restore
	// +kubebuilder:validation:Required
	NodeName string `json:"nodeName"`

	// IncludeArtifacts lists which artifacts to include in response (empty = all)
	// +optional
	IncludeArtifacts []string `json:"includeArtifacts,omitempty"`
}

// SnapshotRestoreState represents restore operation state
type SnapshotRestoreState string

const (
	SnapshotRestoreStatePending     SnapshotRestoreState = "Pending"
	SnapshotRestoreStateDownloading SnapshotRestoreState = "Downloading"
	SnapshotRestoreStateRestoring   SnapshotRestoreState = "Restoring"
	SnapshotRestoreStateCompleted   SnapshotRestoreState = "Completed"
	SnapshotRestoreStateFailed      SnapshotRestoreState = "Failed"
)

// SnapshotRestoreStatus defines restore progress
type SnapshotRestoreStatus struct {
	// State is the current restore state
	// +kubebuilder:default=Pending
	State SnapshotRestoreState `json:"state,omitempty"`

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

// SnapshotRestoreList contains a list of SnapshotRestore
type SnapshotRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnapshotRestore `json:"items"`
}

// ============================================================================
// SnapshotArtifacts - Stores K8s resources captured during snapshot creation
// ============================================================================

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Snapshot",type=string,JSONPath=`.spec.snapshotName`
// +kubebuilder:printcolumn:name="Compositions",type=integer,JSONPath=`.status.compositionCount`
// +kubebuilder:printcolumn:name="ConfigMaps",type=integer,JSONPath=`.status.configMapCount`
// +kubebuilder:printcolumn:name="Secrets",type=integer,JSONPath=`.status.secretCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SnapshotArtifacts stores K8s resources (Compositions, ConfigMaps, Secrets)
// captured during environment snapshot creation. This is a separate resource
// to keep the Snapshot CR lightweight.
type SnapshotArtifacts struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotArtifactsSpec   `json:"spec,omitempty"`
	Status SnapshotArtifactsStatus `json:"status,omitempty"`
}

// SnapshotArtifactsSpec defines the artifacts data
type SnapshotArtifactsSpec struct {
	// SnapshotName is the name of the associated Snapshot
	// +kubebuilder:validation:Required
	SnapshotName string `json:"snapshotName"`

	// Compositions contains serialized Composition resources (YAML, base64 encoded)
	// +optional
	Compositions string `json:"compositions,omitempty"`

	// ConfigMaps contains serialized ConfigMap resources (YAML, base64 encoded)
	// +optional
	ConfigMaps string `json:"configMaps,omitempty"`

	// Secrets contains serialized Secret resources (YAML, base64 encoded)
	// +optional
	Secrets string `json:"secrets,omitempty"`
}

// SnapshotArtifactsStatus tracks counts of stored resources
type SnapshotArtifactsStatus struct {
	// CompositionCount is the number of compositions stored
	CompositionCount int32 `json:"compositionCount,omitempty"`

	// ConfigMapCount is the number of configmaps stored
	ConfigMapCount int32 `json:"configMapCount,omitempty"`

	// SecretCount is the number of secrets stored
	SecretCount int32 `json:"secretCount,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SnapshotArtifactsList contains a list of SnapshotArtifacts
type SnapshotArtifactsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnapshotArtifacts `json:"items"`
}
