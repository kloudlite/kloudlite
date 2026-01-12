// +k8s:deepcopy-gen=package
// +groupName=snapshots.kloudlite.io

// Package v1 contains API Schema definitions for the snapshots v1 API group
//
// This package defines types in the snapshots.kloudlite.io API group:
// - SnapshotStore: Configures storage backends (S3, etc.) for snapshots
// - Snapshot: A namespaced point-in-time snapshot of a btrfs subvolume with artifacts
// - SnapshotRestore: A request to restore a snapshot to a target path
package v1
