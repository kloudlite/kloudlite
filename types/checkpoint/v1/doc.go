// +k8s:deepcopy-gen=package
// +groupName=checkpoints.kloudlite.io

// Package v1 contains API Schema definitions for the checkpoints v1 API group
//
// This package defines types in the checkpoints.kloudlite.io API group:
// - Checkpoint: A point-in-time snapshot of a btrfs subvolume, pushed to OCI registry as a layer
// - CheckpointRestore: A request to restore a checkpoint to a target path
package v1
