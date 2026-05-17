package main

import (
	"context"
	"fmt"
	"path/filepath"

	zap2 "go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func setupWorkspaceHome(logger *zap2.Logger, fs FileSystem) error {
	logger.Info("Setting up workspace home directory",
		zap2.String("path", workspaceHomePath),
		zap2.Int("uid", workspaceUserUID),
		zap2.Int("gid", workspaceUserGID))

	// Create directory if it doesn't exist
	if err := fs.MkdirAll(workspaceHomePath, 0o755); err != nil {
		return fmt.Errorf("failed to create workspace home directory: %w", err)
	}

	// Set ownership to workspace user
	if err := fs.Chown(workspaceHomePath, workspaceUserUID, workspaceUserGID); err != nil {
		return fmt.Errorf("failed to set ownership on workspace home directory: %w", err)
	}

	// Create workspaces subdirectory with correct ownership
	workspacesPath := workspaceHomePath + "/workspaces"
	if err := fs.MkdirAll(workspacesPath, 0o755); err != nil {
		return fmt.Errorf("failed to create workspaces subdirectory: %w", err)
	}

	// Set ownership to workspace user
	if err := fs.Chown(workspacesPath, workspaceUserUID, workspaceUserGID); err != nil {
		return fmt.Errorf("failed to set ownership on workspaces subdirectory: %w", err)
	}

	logger.Info("Successfully set up workspace home directory with workspaces subdirectory",
		zap2.String("path", workspaceHomePath),
		zap2.String("workspacesPath", workspacesPath))

	return nil
}

func setupSSHConfigDirectory(logger *zap2.Logger, fs FileSystem) error {
	logger.Info("Setting up SSH config directory", zap2.String("path", sshConfigPath))

	// Create directory if it doesn't exist
	if err := fs.MkdirAll(sshConfigPath, 0o755); err != nil {
		return fmt.Errorf("failed to create SSH config directory: %w", err)
	}

	logger.Info("Successfully set up SSH config directory", zap2.String("path", sshConfigPath))
	return nil
}

func writeAuthorizedKeys(logger *zap2.Logger, content string, fs FileSystem) error {
	targetPath := filepath.Join(sshConfigPath, authorizedKeysFile)
	tempPath := targetPath + ".tmp"

	// Write to temporary file first (atomic operation)
	if err := fs.WriteFile(tempPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write temporary authorized_keys file: %w", err)
	}

	// Atomically rename temp file to target (atomic on POSIX systems)
	if err := fs.Rename(tempPath, targetPath); err != nil {
		return fmt.Errorf("failed to rename temporary authorized_keys file: %w", err)
	}

	logger.Info("Successfully wrote authorized_keys file",
		zap2.String("path", targetPath),
		zap2.Int("size", len(content)))
	return nil
}

// SSHConfigReconciler watches the ssh-host-keys Secret and writes authorized_keys to the host filesystem
type SSHConfigReconciler struct {
	client.Client
	Logger          *zap2.Logger
	FS              FileSystem
	WorkMachineName string
}

func (r *SSHConfigReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("secret", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling SSH config from Secret")

	// Fetch Secret
	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("Secret deleted or not found")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get Secret", zap2.Error(err))
		return reconcile.Result{}, err
	}

	// Write authorized_keys
	if authorizedKeysBytes, ok := secret.Data["authorized_keys"]; ok {
		if err := writeAuthorizedKeys(logger, string(authorizedKeysBytes), r.FS); err != nil {
			logger.Error("Failed to write authorized_keys", zap2.Error(err))
			return reconcile.Result{}, err
		}
	}

	// Write SSH host keys
	if rsaKeyBytes, ok := secret.Data["ssh_host_rsa_key"]; ok {
		targetPath := filepath.Join(sshConfigPath, "ssh_host_rsa_key")
		tempPath := targetPath + ".tmp"

		// Write to temp file
		if err := r.FS.WriteFile(tempPath, rsaKeyBytes, 0o600); err != nil {
			logger.Error("Failed to write ssh_host_rsa_key temp file", zap2.Error(err))
			return reconcile.Result{}, err
		}

		// Atomic rename
		if err := r.FS.Rename(tempPath, targetPath); err != nil {
			logger.Error("Failed to rename ssh_host_rsa_key", zap2.Error(err))
			return reconcile.Result{}, err
		}

		// Change ownership to kl user (uid 1001) so workspace can use it for SSH
		if err := r.FS.Chown(targetPath, 1001, 1001); err != nil {
			logger.Error("Failed to chown ssh_host_rsa_key", zap2.Error(err))
			return reconcile.Result{}, err
		}
	}

	if rsaPubKeyBytes, ok := secret.Data["ssh_host_rsa_key.pub"]; ok {
		targetPath := filepath.Join(sshConfigPath, "ssh_host_rsa_key.pub")
		tempPath := targetPath + ".tmp"

		// Write to temp file
		if err := r.FS.WriteFile(tempPath, rsaPubKeyBytes, 0o644); err != nil {
			logger.Error("Failed to write ssh_host_rsa_key.pub temp file", zap2.Error(err))
			return reconcile.Result{}, err
		}

		// Atomic rename
		if err := r.FS.Rename(tempPath, targetPath); err != nil {
			logger.Error("Failed to rename ssh_host_rsa_key.pub", zap2.Error(err))
			return reconcile.Result{}, err
		}
	}

	logger.Info("Successfully updated SSH config files")
	return reconcile.Result{}, nil
}

func (r *SSHConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				labels := e.Object.GetLabels()
				return labels != nil &&
					labels["kloudlite.io/ssh-host-keys"] == "true" &&
					labels["kloudlite.io/workmachine"] == r.WorkMachineName
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				labels := e.ObjectNew.GetLabels()
				return labels != nil &&
					labels["kloudlite.io/ssh-host-keys"] == "true" &&
					labels["kloudlite.io/workmachine"] == r.WorkMachineName
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				labels := e.Object.GetLabels()
				return labels != nil &&
					labels["kloudlite.io/ssh-host-keys"] == "true" &&
					labels["kloudlite.io/workmachine"] == r.WorkMachineName
			},
		}).
		Complete(r)
}
