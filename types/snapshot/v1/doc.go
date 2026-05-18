// +k8s:deepcopy-gen=package
// +groupName=snapshots.kloudlite.io

// Package v1 contains API Schema definitions for the snapshots v1 API group
//
// This package defines types in the snapshots.kloudlite.io API group:
// - Snapshot: A namespaced point-in-time snapshot of a btrfs subvolume with artifacts
// - SnapshotRequest: A request to create a snapshot on a specific node
// - SnapshotRestore: A request to restore a snapshot to a target path
// - SnapshotArtifacts: Stores K8s resources captured during snapshot creation
package v1
