package workspace

import (
	"fmt"

	workspacev1 "github.com/kloudlite/kloudlite/types/workspace/v1"
	"go.uber.org/zap"
)

const workspaceStoragePath = "/var/lib/kloudlite/storage/workspaces"

// ensureBtrfsSubvolume creates a btrfs subvolume for the workspace if it doesn't already exist.
func (r *WorkspaceReconciler) ensureBtrfsSubvolume(workspace *workspacev1.Workspace, logger *zap.Logger) error {
	workspaceDir := fmt.Sprintf("%s/%s", workspaceStoragePath, workspace.Name)

	// Check if btrfs subvolume already exists
	checkScript := fmt.Sprintf("btrfs subvolume show %s > /dev/null 2>&1", workspaceDir)
	if _, err := r.CmdExec.Execute(checkScript); err == nil {
		logger.Debug("Workspace btrfs subvolume already exists", zap.String("path", workspaceDir))
		return nil
	}

	// Create new btrfs subvolume
	logger.Info("Creating workspace btrfs subvolume", zap.String("path", workspaceDir))
	createScript := fmt.Sprintf("btrfs subvolume create %s && chown 1001:1001 %s", workspaceDir, workspaceDir)

	if output, err := r.CmdExec.Execute(createScript); err != nil {
		return fmt.Errorf("failed to create btrfs subvolume at %s: %w (output: %s)", workspaceDir, err, string(output))
	}

	logger.Info("Successfully created workspace btrfs subvolume", zap.String("path", workspaceDir))
	return nil
}

// deleteBtrfsSubvolume deletes the btrfs subvolume for the workspace.
func (r *WorkspaceReconciler) deleteBtrfsSubvolume(workspace *workspacev1.Workspace, logger *zap.Logger) error {
	workspaceDir := fmt.Sprintf("%s/%s", workspaceStoragePath, workspace.Name)

	logger.Info("Removing workspace btrfs subvolume", zap.String("path", workspaceDir))

	deleteScript := fmt.Sprintf("btrfs subvolume delete %s", workspaceDir)
	output, err := r.CmdExec.Execute(deleteScript)
	if err != nil {
		// Check if path exists - if not, nothing to clean up
		checkExistsScript := fmt.Sprintf("test -e %s", workspaceDir)
		if _, existsErr := r.CmdExec.Execute(checkExistsScript); existsErr != nil {
			logger.Info("Workspace subvolume doesn't exist, skipping cleanup", zap.String("path", workspaceDir))
			return nil
		}
		return fmt.Errorf("failed to delete btrfs subvolume at %s: %w (output: %s)", workspaceDir, err, string(output))
	}

	logger.Info("Successfully deleted workspace btrfs subvolume", zap.String("path", workspaceDir))
	return nil
}
