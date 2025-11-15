package workspace

import (
	"context"
	"fmt"
	"time"

	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// handleCloning manages the workspace directory cloning process
// This function orchestrates the entire cloning workflow through various phases
func (r *WorkspaceReconciler) handleCloning(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Handling workspace cloning",
		zap.String("workspace", workspace.Name),
		zap.String("copyFrom", workspace.Spec.CopyFrom))

	// Initialize cloning status if not set
	if workspace.Status.CloningStatus == nil {
		logger.Info("Initializing cloning status")
		workspace.Status.CloningStatus = &workspacev1.WorkspaceCloningStatus{
			Phase:               workspacev1.CloningPhasePending,
			Message:             "Cloning initialized",
			SourceWorkspaceName: workspace.Spec.CopyFrom,
			StartTime:           fn.Ptr(metav1.Now()),
		}
		if err := r.Status().Update(ctx, workspace); err != nil {
			logger.Error("Failed to initialize cloning status", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	status := workspace.Status.CloningStatus

	// Handle based on current phase
	switch status.Phase {
	case workspacev1.CloningPhasePending:
		return r.handleCloningPending(ctx, workspace, logger)

	case workspacev1.CloningPhaseSuspending:
		return r.handleCloningSuspending(ctx, workspace, logger)

	case workspacev1.CloningPhaseCreatingCopyJob:
		return r.handleCloningCreatingCopyJob(ctx, workspace, logger)

	case workspacev1.CloningPhaseWaitingForCopyCompletion:
		return r.handleCloningWaitingForCopyCompletion(ctx, workspace, logger)

	case workspacev1.CloningPhaseVerifyingCopy:
		return r.handleCloningVerifyingCopy(ctx, workspace, logger)

	case workspacev1.CloningPhaseResuming:
		return r.handleCloningResuming(ctx, workspace, logger)

	case workspacev1.CloningPhaseCompleted:
		return r.handleCloningCompleted(ctx, workspace, logger)

	case workspacev1.CloningPhaseFailed:
		// Cloning failed, resume source workspace if needed
		return r.handleCloningFailed(ctx, workspace, logger)

	default:
		logger.Error("Unknown cloning phase", zap.String("phase", string(status.Phase)))
		return reconcile.Result{}, nil
	}
}

// handleCloningPending validates the source workspace and moves to Suspending phase
func (r *WorkspaceReconciler) handleCloningPending(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: Pending - Validating source workspace")

	// Fetch source workspace
	sourceWorkspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: workspace.Spec.CopyFrom}, sourceWorkspace); err != nil {
		logger.Error("Failed to get source workspace", zap.Error(err))
		workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseFailed
		workspace.Status.CloningStatus.ErrorMessage = fmt.Sprintf("Source workspace not found: %v", err)
		workspace.Status.CloningStatus.CompletionTime = fn.Ptr(metav1.Now())
		if err := r.Status().Update(ctx, workspace); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// Validate source workspace state
	if sourceWorkspace.DeletionTimestamp != nil {
		logger.Error("Source workspace is being deleted")
		workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseFailed
		workspace.Status.CloningStatus.ErrorMessage = "Source workspace is being deleted"
		workspace.Status.CloningStatus.CompletionTime = fn.Ptr(metav1.Now())
		if err := r.Status().Update(ctx, workspace); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// Store source workspace details
	workspace.Status.CloningStatus.SourceWorkspaceName = sourceWorkspace.Name
	workspace.Status.CloningStatus.SourceWorkmachineName = sourceWorkspace.Spec.WorkmachineName
	workspace.Status.CloningStatus.SourceFolderName = sourceWorkspace.Spec.FolderName

	// Move to Suspending phase
	workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseSuspending
	workspace.Status.CloningStatus.Message = "Suspending source workspace"

	if err := r.Status().Update(ctx, workspace); err != nil {
		logger.Error("Failed to update cloning status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Moving to Suspending phase")
	return reconcile.Result{Requeue: true}, nil
}

// handleCloningSuspending suspends the source workspace pod
func (r *WorkspaceReconciler) handleCloningSuspending(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: Suspending - Suspending source workspace pod")

	// Fetch source workspace
	sourceWorkspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: workspace.Spec.CopyFrom}, sourceWorkspace); err != nil {
		logger.Error("Failed to get source workspace", zap.Error(err))
		workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseFailed
		workspace.Status.CloningStatus.ErrorMessage = fmt.Sprintf("Failed to get source workspace: %v", err)
		if err := r.Status().Update(ctx, workspace); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// Set source cloning status
	if sourceWorkspace.Status.SourceCloningStatus == nil {
		logger.Info("Setting source cloning status on source workspace")
		sourceWorkspace.Status.SourceCloningStatus = &workspacev1.WorkspaceSourceCloningStatus{
			Phase:                 workspacev1.SourceCloningPhaseSuspending,
			Message:               "Being used as clone source, suspending workspace",
			TargetWorkspaceName:   workspace.Name,
			TargetWorkmachineName: workspace.Spec.WorkmachineName,
			StartTime:             fn.Ptr(metav1.Now()),
		}
		if err := r.Status().Update(ctx, sourceWorkspace); err != nil {
			logger.Error("Failed to update source workspace status", zap.Error(err))
			return reconcile.Result{}, err
		}
	}

	// Suspend source workspace by deleting its pod
	if err := r.suspendWorkspacePod(ctx, sourceWorkspace, logger); err != nil {
		logger.Error("Failed to suspend source workspace", zap.Error(err))
		workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseFailed
		workspace.Status.CloningStatus.ErrorMessage = fmt.Sprintf("Failed to suspend source workspace: %v", err)
		if err := r.Status().Update(ctx, workspace); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// Move to CreatingCopyJob phase
	workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseCreatingCopyJob
	workspace.Status.CloningStatus.Message = "Creating directory copy jobs"

	// Update source workspace phase
	sourceWorkspace.Status.SourceCloningStatus.Phase = workspacev1.SourceCloningPhaseCopying
	sourceWorkspace.Status.SourceCloningStatus.Message = "Directory being copied"

	if err := r.Status().Update(ctx, workspace); err != nil {
		logger.Error("Failed to update cloning status", zap.Error(err))
		return reconcile.Result{}, err
	}

	if err := r.Status().Update(ctx, sourceWorkspace); err != nil {
		logger.Error("Failed to update source workspace status", zap.Error(err))
	}

	logger.Info("Source workspace suspended, moving to CreatingCopyJob phase")
	return reconcile.Result{Requeue: true}, nil
}

// handleCloningCreatingCopyJob creates sender and receiver jobs for directory copying
func (r *WorkspaceReconciler) handleCloningCreatingCopyJob(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: CreatingCopyJob - Creating sender and receiver jobs")

	// Fetch source workspace
	sourceWorkspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: workspace.Spec.CopyFrom}, sourceWorkspace); err != nil {
		logger.Error("Failed to get source workspace", zap.Error(err))
		workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseFailed
		workspace.Status.CloningStatus.ErrorMessage = fmt.Sprintf("Failed to get source workspace: %v", err)
		if err := r.Status().Update(ctx, workspace); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// Initialize directory copier
	copier := &WorkspaceDirectoryCopier{
		Client: r.Client,
		Logger: logger,
	}

	// Initialize copy job status
	if workspace.Status.CloningStatus.CopyJobStatus == nil {
		workspace.Status.CloningStatus.CopyJobStatus = &workspacev1.DirectoryCopyJobStatus{
			StartTime: fn.Ptr(metav1.Now()),
		}
	}

	copyStatus := workspace.Status.CloningStatus.CopyJobStatus

	// Create sender job if not already created
	if copyStatus.SenderJobName == "" {
		logger.Info("Creating sender job")
		senderJob, err := copier.createSenderJob(ctx, workspace, sourceWorkspace, logger)
		if err != nil {
			logger.Error("Failed to create sender job", zap.Error(err))
			workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseFailed
			workspace.Status.CloningStatus.ErrorMessage = fmt.Sprintf("Failed to create sender job: %v", err)
			if err := r.Status().Update(ctx, workspace); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}

		copyStatus.SenderJobName = senderJob.Name
		copyStatus.Started = true

		if err := r.Status().Update(ctx, workspace); err != nil {
			logger.Error("Failed to update cloning status with sender job", zap.Error(err))
			return reconcile.Result{}, err
		}

		logger.Info("Sender job created", zap.String("jobName", senderJob.Name))
	}

	// Wait for sender pod to get IP
	if copyStatus.SenderPodIP == "" {
		logger.Info("Waiting for sender pod to get IP")

		namespace, err := copier.getJobNamespace(sourceWorkspace)
		if err != nil {
			logger.Error("Failed to get job namespace", zap.Error(err))
			workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseFailed
			workspace.Status.CloningStatus.ErrorMessage = fmt.Sprintf("Failed to get job namespace: %v", err)
			if err := r.Status().Update(ctx, workspace); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}

		senderPodIP, err := copier.waitForSenderPodReady(ctx, copyStatus.SenderJobName, namespace, logger)
		if err != nil {
			logger.Error("Failed to get sender pod IP", zap.Error(err))
			workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseFailed
			workspace.Status.CloningStatus.ErrorMessage = fmt.Sprintf("Failed to get sender pod IP: %v", err)
			if err := r.Status().Update(ctx, workspace); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}

		copyStatus.SenderPodIP = senderPodIP
		copyStatus.Message = fmt.Sprintf("Sender pod ready at %s", senderPodIP)

		if err := r.Status().Update(ctx, workspace); err != nil {
			logger.Error("Failed to update cloning status with sender IP", zap.Error(err))
			return reconcile.Result{}, err
		}

		logger.Info("Sender pod ready", zap.String("ip", senderPodIP))
	}

	// Create receiver job if not already created
	if copyStatus.ReceiverJobName == "" {
		logger.Info("Creating receiver job")
		receiverJob, err := copier.createReceiverJob(ctx, workspace, copyStatus.SenderPodIP, logger)
		if err != nil {
			logger.Error("Failed to create receiver job", zap.Error(err))
			workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseFailed
			workspace.Status.CloningStatus.ErrorMessage = fmt.Sprintf("Failed to create receiver job: %v", err)
			if err := r.Status().Update(ctx, workspace); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}

		copyStatus.ReceiverJobName = receiverJob.Name
		copyStatus.Message = "Receiver job created, copying directory"

		if err := r.Status().Update(ctx, workspace); err != nil {
			logger.Error("Failed to update cloning status with receiver job", zap.Error(err))
			return reconcile.Result{}, err
		}

		logger.Info("Receiver job created", zap.String("jobName", receiverJob.Name))
	}

	// Move to WaitingForCopyCompletion phase
	workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseWaitingForCopyCompletion
	workspace.Status.CloningStatus.Message = "Waiting for directory copy to complete"

	if err := r.Status().Update(ctx, workspace); err != nil {
		logger.Error("Failed to update cloning status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Copy jobs created, moving to WaitingForCopyCompletion phase")
	return reconcile.Result{Requeue: true}, nil
}

// handleCloningWaitingForCopyCompletion waits for the receiver job to complete
func (r *WorkspaceReconciler) handleCloningWaitingForCopyCompletion(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: WaitingForCopyCompletion - Checking copy job status")

	copyStatus := workspace.Status.CloningStatus.CopyJobStatus
	if copyStatus == nil || copyStatus.ReceiverJobName == "" {
		logger.Error("Copy job status not found")
		workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseFailed
		workspace.Status.CloningStatus.ErrorMessage = "Copy job status not found"
		if err := r.Status().Update(ctx, workspace); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// Initialize directory copier
	copier := &WorkspaceDirectoryCopier{
		Client: r.Client,
		Logger: logger,
	}

	// Get job namespace
	namespace, err := copier.getJobNamespace(workspace)
	if err != nil {
		logger.Error("Failed to get job namespace", zap.Error(err))
		workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseFailed
		workspace.Status.CloningStatus.ErrorMessage = fmt.Sprintf("Failed to get job namespace: %v", err)
		if err := r.Status().Update(ctx, workspace); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// Check receiver job status
	completed, failed, err := copier.getDirectoryCopyStatus(ctx, copyStatus.ReceiverJobName, namespace, logger)
	if err != nil {
		logger.Error("Failed to get copy status", zap.Error(err))
		// Requeue to retry
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	if failed {
		logger.Error("Copy job failed")
		copyStatus.Failed = true
		copyStatus.Message = "Directory copy failed"
		copyStatus.CompletionTime = fn.Ptr(metav1.Now())

		workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseFailed
		workspace.Status.CloningStatus.ErrorMessage = "Directory copy job failed"
		workspace.Status.CloningStatus.CompletionTime = fn.Ptr(metav1.Now())

		if err := r.Status().Update(ctx, workspace); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	if !completed {
		logger.Info("Copy job still running, will check again")
		copyStatus.Message = "Directory copy in progress"
		if err := r.Status().Update(ctx, workspace); err != nil {
			return reconcile.Result{}, err
		}
		// Requeue after 10 seconds to check again
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Copy completed successfully
	logger.Info("Copy job completed successfully")
	copyStatus.Completed = true
	copyStatus.Message = "Directory copy completed successfully"
	copyStatus.CompletionTime = fn.Ptr(metav1.Now())

	workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseVerifyingCopy
	workspace.Status.CloningStatus.Message = "Verifying directory copy"

	if err := r.Status().Update(ctx, workspace); err != nil {
		logger.Error("Failed to update cloning status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Moving to VerifyingCopy phase")
	return reconcile.Result{Requeue: true}, nil
}

// handleCloningVerifyingCopy verifies the directory was copied successfully
func (r *WorkspaceReconciler) handleCloningVerifyingCopy(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: VerifyingCopy - Verifying directory copy")

	// For now, we trust the receiver job's success status
	// In the future, we could add additional verification (e.g., check directory exists, compare file counts)

	// Move to Resuming phase
	workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseResuming
	workspace.Status.CloningStatus.Message = "Resuming source workspace"

	if err := r.Status().Update(ctx, workspace); err != nil {
		logger.Error("Failed to update cloning status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Verification passed, moving to Resuming phase")
	return reconcile.Result{Requeue: true}, nil
}

// handleCloningResuming resumes the source workspace pod
func (r *WorkspaceReconciler) handleCloningResuming(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: Resuming - Resuming source workspace")

	// Fetch source workspace
	sourceWorkspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: workspace.Spec.CopyFrom}, sourceWorkspace); err != nil {
		logger.Warn("Failed to get source workspace for resume", zap.Error(err))
		// Don't fail cloning if source workspace is gone, just log it
	} else {
		// Update source workspace status
		if sourceWorkspace.Status.SourceCloningStatus != nil {
			sourceWorkspace.Status.SourceCloningStatus.Phase = workspacev1.SourceCloningPhaseResuming
			sourceWorkspace.Status.SourceCloningStatus.Message = "Cloning completed, resuming workspace"
			sourceWorkspace.Status.SourceCloningStatus.CompletionTime = fn.Ptr(metav1.Now())

			if err := r.Status().Update(ctx, sourceWorkspace); err != nil {
				logger.Warn("Failed to update source workspace status", zap.Error(err))
			}
		}

		// Resume source workspace by triggering reconciliation
		// The workspace controller will recreate the pod if status is "active"
		// We just need to clear the SourceCloningStatus
		if sourceWorkspace.Status.SourceCloningStatus != nil {
			sourceWorkspace.Status.SourceCloningStatus = nil
			if err := r.Status().Update(ctx, sourceWorkspace); err != nil {
				logger.Warn("Failed to clear source cloning status", zap.Error(err))
			}
		}
	}

	// Cleanup copy jobs
	copier := &WorkspaceDirectoryCopier{
		Client: r.Client,
		Logger: logger,
	}

	namespace, err := copier.getJobNamespace(workspace)
	if err != nil {
		logger.Warn("Failed to get job namespace for cleanup", zap.Error(err))
	} else {
		copyStatus := workspace.Status.CloningStatus.CopyJobStatus
		if copyStatus != nil {
			if err := copier.cleanupCopyJobs(
				ctx,
				copyStatus.SenderJobName,
				copyStatus.ReceiverJobName,
				namespace,
				logger,
			); err != nil {
				logger.Warn("Failed to cleanup copy jobs", zap.Error(err))
			}
		}
	}

	// Move to Completed phase
	workspace.Status.CloningStatus.Phase = workspacev1.CloningPhaseCompleted
	workspace.Status.CloningStatus.Message = "Cloning completed successfully"
	workspace.Status.CloningStatus.CompletionTime = fn.Ptr(metav1.Now())

	if err := r.Status().Update(ctx, workspace); err != nil {
		logger.Error("Failed to update cloning status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Source workspace resumed, moving to Completed phase")
	return reconcile.Result{Requeue: true}, nil
}

// handleCloningCompleted clears the copyFrom field and cloning status
func (r *WorkspaceReconciler) handleCloningCompleted(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: Completed - Clearing copyFrom field")

	// Clear copyFrom field to mark cloning as complete
	workspace.Spec.CopyFrom = ""

	// Add completion condition
	workspace.Status.Conditions = append(workspace.Status.Conditions, metav1.Condition{
		Type:               "Cloned",
		Status:             metav1.ConditionTrue,
		ObservedGeneration: workspace.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "CloningCompleted",
		Message:            fmt.Sprintf("Successfully cloned from workspace %s", workspace.Status.CloningStatus.SourceWorkspaceName),
	})

	// Clear cloning status (keep it for a while for debugging)
	// workspace.Status.CloningStatus = nil

	if err := r.Update(ctx, workspace); err != nil {
		logger.Error("Failed to clear copyFrom field", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Cloning completed successfully, copyFrom field cleared")

	// Requeue to start normal workspace reconciliation
	return reconcile.Result{Requeue: true}, nil
}

// handleCloningFailed handles failed cloning by resuming source workspace
func (r *WorkspaceReconciler) handleCloningFailed(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Error("Phase: Failed - Cloning failed, attempting to resume source workspace")

	// Try to resume source workspace
	if workspace.Spec.CopyFrom != "" {
		sourceWorkspace := &workspacev1.Workspace{}
		if err := r.Get(ctx, client.ObjectKey{Name: workspace.Spec.CopyFrom}, sourceWorkspace); err == nil {
			// Clear source cloning status to allow it to resume
			if sourceWorkspace.Status.SourceCloningStatus != nil {
				sourceWorkspace.Status.SourceCloningStatus = nil
				if err := r.Status().Update(ctx, sourceWorkspace); err != nil {
					logger.Warn("Failed to clear source cloning status", zap.Error(err))
				}
			}
		}
	}

	// Add failure condition
	workspace.Status.Conditions = append(workspace.Status.Conditions, metav1.Condition{
		Type:               "Cloned",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: workspace.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "CloningFailed",
		Message:            workspace.Status.CloningStatus.ErrorMessage,
	})

	if err := r.Status().Update(ctx, workspace); err != nil {
		logger.Error("Failed to update status with failure condition", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Don't requeue, let user fix the issue and retry manually
	logger.Info("Cloning failed, source workspace resumed")
	return reconcile.Result{}, nil
}

// suspendWorkspacePod deletes the workspace pod to suspend it
func (r *WorkspaceReconciler) suspendWorkspacePod(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) error {
	logger.Info("Suspending workspace pod",
		zap.String("workspace", workspace.Name),
		zap.String("podName", workspace.Status.PodName))

	// Delete the workspace pod if it exists
	if workspace.Status.PodName != "" {
		if err := r.deleteWorkspacePod(ctx, workspace, logger); err != nil {
			return fmt.Errorf("failed to delete workspace pod: %w", err)
		}
		logger.Info("Workspace pod deleted successfully")
	}

	return nil
}
