package oci

import (
	"time"
)

// SnapshotMetadata contains the metadata stored alongside each snapshot layer
type SnapshotMetadata struct {
	// Name is the original snapshot name
	Name string `json:"name"`

	// Spec contains the snapshot spec fields
	Spec SnapshotMetadataSpec `json:"spec"`

	// Status contains the snapshot status fields
	Status SnapshotMetadataStatus `json:"status"`

	// Resources contains K8s resource metadata for environment snapshots
	// This allows resource restoration without reading files from the snapshot
	Resources *ResourceMetadata `json:"resources,omitempty"`
}

// SnapshotMetadataSpec mirrors relevant fields from SnapshotSpec
type SnapshotMetadataSpec struct {
	Description       string                   `json:"description,omitempty"`
	OwnedBy           string                   `json:"ownedBy"`
	IncludeMetadata   bool                     `json:"includeMetadata"`
	ParentSnapshotRef *ParentSnapshotReference `json:"parentSnapshotRef,omitempty"`
	EnvironmentRef    *EnvironmentReference    `json:"environmentRef,omitempty"`
	WorkspaceRef      *WorkspaceReference      `json:"workspaceRef,omitempty"`
}

// ParentSnapshotReference identifies the parent snapshot
type ParentSnapshotReference struct {
	Name       string     `json:"name"`
	RestoredAt *time.Time `json:"restoredAt,omitempty"`
}

// EnvironmentReference is a reference to an environment
type EnvironmentReference struct {
	Name string `json:"name"`
}

// WorkspaceReference is a reference to a workspace
type WorkspaceReference struct {
	Name            string `json:"name"`
	WorkmachineName string `json:"workmachineName"`
}

// SnapshotMetadataStatus mirrors relevant fields from SnapshotStatus
type SnapshotMetadataStatus struct {
	SnapshotType    string     `json:"snapshotType,omitempty"`
	TargetName      string     `json:"targetName,omitempty"`
	SizeBytes       int64      `json:"sizeBytes,omitempty"`
	SizeHuman       string     `json:"sizeHuman,omitempty"`
	CreatedAt       *time.Time `json:"createdAt,omitempty"`
	WorkMachineName string     `json:"workMachineName,omitempty"`
	WorkspaceName   string     `json:"workspaceName,omitempty"`
}

// ResourceMetadata contains K8s resource JSON for environment snapshots
// This is stored in metadata.json alongside data.tar.gz in the OCI layer
type ResourceMetadata struct {
	ConfigMaps   string `json:"configMaps,omitempty"`
	Secrets      string `json:"secrets,omitempty"`
	Deployments  string `json:"deployments,omitempty"`
	Services     string `json:"services,omitempty"`
	StatefulSets string `json:"statefulSets,omitempty"`
	Compositions string `json:"compositions,omitempty"`
}

// ImageConfig is the OCI image config for snapshot images
type ImageConfig struct {
	ImageType string `json:"imageType"`
	Version   string `json:"version"`
}

// PushOptions contains options for pushing a snapshot to registry
type PushOptions struct {
	// RegistryURL is the registry base URL (e.g., "image-registry:5000")
	RegistryURL string

	// Repository is the image repository (e.g., "snapshots/username")
	Repository string

	// Tag is the image tag
	Tag string

	// SnapshotPath is the path to the btrfs snapshot
	SnapshotPath string

	// ParentSnapshotPath is the path to parent snapshot for incremental send
	ParentSnapshotPath string

	// Metadata is the snapshot metadata to include in the layer
	Metadata *SnapshotMetadata

	// ParentLayers are existing layer digests to include from parent image
	ParentLayers []string

	// Insecure allows HTTP registry connections
	Insecure bool
}

// PushResult contains the result of a push operation
type PushResult struct {
	// ImageRef is the full image reference
	ImageRef string

	// Digest is the manifest digest
	Digest string

	// LayerDigests are all layer digests in order
	LayerDigests []string

	// CompressedSize is the total compressed size
	CompressedSize int64
}

// PullOptions contains options for pulling a snapshot from registry
type PullOptions struct {
	// RegistryURL is the registry base URL
	RegistryURL string

	// Repository is the image repository
	Repository string

	// Tag is the image tag
	Tag string

	// TargetDir is where to receive the btrfs snapshots
	TargetDir string

	// Insecure allows HTTP registry connections
	Insecure bool
}

// PullResult contains the result of a pull operation
type PullResult struct {
	// Snapshots contains metadata for each pulled snapshot
	Snapshots []SnapshotMetadata

	// SnapshotPaths maps snapshot names to their local paths
	SnapshotPaths map[string]string
}
