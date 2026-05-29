package workmachinenodemanager

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	checkpointv1 "github.com/kloudlite/kloudlite/types/checkpoint/v1"
	environmentv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	zap2 "go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// StorageGarbageCollector periodically cleans up orphaned storage directories
type StorageGarbageCollector struct {
	// Reader is used to read resources directly from API server (bypasses cache)
	// This avoids needing watch permissions which would require cache sync
	Reader      client.Reader
	Logger      *zap2.Logger
	HostCmdExec CommandExecutor
	Interval    time.Duration
}

// Run starts the garbage collector loop
func (gc *StorageGarbageCollector) Run(ctx context.Context) {
	ticker := time.NewTicker(gc.Interval)
	defer ticker.Stop()

	// Run immediately on start
	gc.collectGarbage(ctx)

	for {
		select {
		case <-ctx.Done():
			gc.Logger.Info("Storage garbage collector stopped")
			return
		case <-ticker.C:
			gc.collectGarbage(ctx)
		}
	}
}

// collectGarbage performs the actual garbage collection
func (gc *StorageGarbageCollector) collectGarbage(ctx context.Context) {
	gc.Logger.Info("Running storage garbage collection")

	// Clean up orphaned environment directories
	gc.cleanupOrphanedEnvironments(ctx)

	// Clean up orphaned checkpoint cache directories
	gc.cleanupOrphanedCheckpointCache(ctx)
}

// cleanupOrphanedEnvironments removes environment directories that don't have a corresponding Environment CR
func (gc *StorageGarbageCollector) cleanupOrphanedEnvironments(ctx context.Context) {
	envStoragePath := "/var/lib/kloudlite/storage/environments"

	// List directories on disk
	listCmd := fmt.Sprintf("ls -1 %s 2>/dev/null || true", envStoragePath)
	output, err := gc.HostCmdExec.Execute(listCmd)
	if err != nil {
		gc.Logger.Error("Failed to list environment directories", zap2.Error(err))
		return
	}

	dirs := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(dirs) == 0 || (len(dirs) == 1 && dirs[0] == "") {
		return
	}

	// Get all existing environments
	envList := &environmentv1.EnvironmentList{}
	if err := gc.Reader.List(ctx, envList); err != nil {
		gc.Logger.Error("Failed to list environments", zap2.Error(err))
		return
	}

	// Build set of valid environment namespaces
	validNamespaces := make(map[string]bool)
	for _, env := range envList.Items {
		validNamespaces[env.Spec.TargetNamespace] = true
	}

	// Check each directory
	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		// Skip non-environment directories (e.g., wm-* workspace directories)
		if !strings.HasPrefix(dir, "env-") {
			continue
		}

		// Check if this directory has a corresponding Environment
		if !validNamespaces[dir] {
			gc.Logger.Info("Found orphaned environment directory", zap2.String("dir", dir))

			// Delete the btrfs subvolume
			dirPath := filepath.Join(envStoragePath, dir)
			deleteCmd := fmt.Sprintf("btrfs subvolume delete %s 2>/dev/null || rm -rf %s", dirPath, dirPath)
			if _, err := gc.HostCmdExec.Execute(deleteCmd); err != nil {
				gc.Logger.Error("Failed to delete orphaned environment directory",
					zap2.String("dir", dir),
					zap2.Error(err))
			} else {
				gc.Logger.Info("Deleted orphaned environment directory", zap2.String("dir", dir))
			}
		}
	}
}

// cleanupOrphanedCheckpointCache removes checkpoint cache directories that don't have a corresponding Checkpoint CR
func (gc *StorageGarbageCollector) cleanupOrphanedCheckpointCache(ctx context.Context) {
	checkpointCachePath := checkpointStoragePath

	// List directories on disk
	listCmd := fmt.Sprintf("ls -1 %s 2>/dev/null || true", checkpointCachePath)
	output, err := gc.HostCmdExec.Execute(listCmd)
	if err != nil {
		gc.Logger.Error("Failed to list checkpoint cache directories", zap2.Error(err))
		return
	}

	dirs := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(dirs) == 0 || (len(dirs) == 1 && dirs[0] == "") {
		return
	}

	// Get all existing checkpoints
	checkpointList := &checkpointv1.CheckpointList{}
	if err := gc.Reader.List(ctx, checkpointList); err != nil {
		gc.Logger.Error("Failed to list checkpoints", zap2.Error(err))
		return
	}

	// Build set of valid checkpoint names
	validCheckpoints := make(map[string]bool)
	for _, cp := range checkpointList.Items {
		validCheckpoints[cp.Name] = true
	}

	// Check each directory
	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		// Check if this directory has a corresponding Checkpoint
		if !validCheckpoints[dir] {
			gc.Logger.Info("Found orphaned checkpoint cache directory", zap2.String("dir", dir))

			// Delete the btrfs subvolume
			dirPath := filepath.Join(checkpointCachePath, dir)
			deleteCmd := fmt.Sprintf("btrfs subvolume delete %s 2>/dev/null || rm -rf %s", dirPath, dirPath)
			if _, err := gc.HostCmdExec.Execute(deleteCmd); err != nil {
				gc.Logger.Error("Failed to delete orphaned checkpoint cache directory",
					zap2.String("dir", dir),
					zap2.Error(err))
			} else {
				gc.Logger.Info("Deleted orphaned checkpoint cache directory", zap2.String("dir", dir))
			}
		}
	}
}
