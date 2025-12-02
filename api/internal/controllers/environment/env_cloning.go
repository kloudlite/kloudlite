package environment

import (
	"context"
	"fmt"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// handleCloning handles cloning resources from a source environment including PVCs
func (r *EnvironmentReconciler) handleCloning(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (reconcile.Result, error) {
	sourceName := environment.Spec.CloneFrom

	logger.Info("Starting environment cloning process",
		zap.String("target", environment.Name),
		zap.String("source", sourceName))

	// Initialize cloning status if not already set
	if environment.Status.CloningStatus == nil {
		now := metav1.Now()
		environment.Status.CloningStatus = &environmentsv1.CloningStatus{
			Phase:     environmentsv1.CloningPhasePending,
			Message:   "Initializing cloning process",
			StartTime: &now,
		}
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
			return nil
		}, logger); err != nil {
			logger.Error("Failed to initialize cloning status", zap.Error(err))
		}
	}

	// Validate source environment exists and is accessible
	sourceEnv := &environmentsv1.Environment{}
	err := r.Get(ctx, client.ObjectKey{Name: sourceName}, sourceEnv)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error("Source environment not found", zap.String("source", sourceName))
			r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed, "Source environment not found: "+sourceName, logger)
			if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateError, "Source environment not found: "+sourceName, logger); err != nil {
				logger.Error("Failed to update environment status after retries", zap.Error(err))
			}
			return reconcile.Result{}, fmt.Errorf("source environment '%s' not found", sourceName)
		}
		logger.Error("Failed to get source environment", zap.Error(err))
		return reconcile.Result{}, fmt.Errorf("failed to access source environment '%s': %w", sourceName, err)
	}

	// Validate source environment state
	if sourceEnv.Status.State == environmentsv1.EnvironmentStateDeleting || sourceEnv.Status.State == environmentsv1.EnvironmentStateError {
		logger.Error("Source environment is not in a clonable state",
			zap.String("source", sourceName),
			zap.String("sourceState", string(sourceEnv.Status.State)))
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed,
			fmt.Sprintf("Source environment is in %s state", sourceEnv.Status.State), logger)
		if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateError,
			fmt.Sprintf("Source environment '%s' is in %s state and cannot be cloned", sourceName, sourceEnv.Status.State), logger); err != nil {
			logger.Error("Failed to update environment status after retries", zap.Error(err))
		}
		return reconcile.Result{}, fmt.Errorf("source environment '%s' is in %s state", sourceName, sourceEnv.Status.State)
	}

	sourceNamespace := sourceEnv.Spec.TargetNamespace
	targetNamespace := environment.Spec.TargetNamespace

	logger.Info("Cloning environment resources",
		zap.String("source", sourceName),
		zap.String("sourceNamespace", sourceNamespace),
		zap.String("targetNamespace", targetNamespace))

	// Phase 1: Suspend source environment if it's active
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhasePending {
		if sourceEnv.Spec.Activated {
			logger.Info("Suspending source environment for safe cloning")
			r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseSuspending, "Suspending source environment", logger)

			// Set source cloning status
			now := metav1.Now()
			sourceEnv.Status.SourceCloningStatus = &environmentsv1.SourceCloningStatus{
				TargetEnvironmentName: environment.Name,
				Phase:                 environmentsv1.SourceCloningPhaseSuspended,
				Message:               fmt.Sprintf("Environment suspended for cloning to %s", environment.Name),
				StartTime:             &now,
			}
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, sourceEnv, func() error {
				return nil
			}, logger); err != nil {
				logger.Error("Failed to update source cloning status", zap.Error(err))
			}

			// Scale down source environment
			if err := r.suspendEnvironment(ctx, sourceEnv, logger); err != nil {
				logger.Error("Failed to suspend source environment", zap.Error(err))
				r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed, "Failed to suspend source environment", logger)
				return reconcile.Result{RequeueAfter: 10 * time.Second}, err
			}
			logger.Info("Source environment suspended successfully")
		}

		// Move to next phase
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseCloningResources, "Starting resource cloning", logger)
		return reconcile.Result{Requeue: true}, nil
	}

	// Note: TargetNamespace is always set by the mutation webhook for the cloned environment.
	// Controller creates the actual Kubernetes namespace resource if it doesn't exist.

	// Create target namespace if it doesn't exist
	namespace := &corev1.Namespace{}
	err = r.Get(ctx, client.ObjectKey{Name: targetNamespace}, namespace)
	if apierrors.IsNotFound(err) {
		if err := r.createNamespaceForCloning(ctx, environment, sourceName, logger); err != nil {
			r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed, fmt.Sprintf("Failed to create namespace: %v", err), logger)
			if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateError,
				fmt.Sprintf("Failed to create namespace: %v", err), logger); err != nil {
				logger.Error("Failed to update environment status after retries", zap.Error(err))
			}
			return reconcile.Result{RequeueAfter: 30 * time.Second}, err
		}
	}

	// Clone ConfigMaps with label "kloudlite.io/resource-type: environment-config"
	logger.Info("Cloning ConfigMaps from source environment")
	configMapList := &corev1.ConfigMapList{}
	err = r.List(ctx, configMapList,
		client.InNamespace(sourceNamespace),
		client.MatchingLabels{"kloudlite.io/resource-type": "environment-config"})
	if err != nil {
		logger.Error("Failed to list source configmaps", zap.Error(err))
		if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateError,
			fmt.Sprintf("Failed to list ConfigMaps from source environment '%s': %v", sourceName, err), logger); err != nil {
			logger.Error("Failed to update environment status after retries", zap.Error(err))
		}
		return reconcile.Result{}, fmt.Errorf("failed to list source ConfigMaps: %w", err)
	}

	clonedConfigMaps := 0
	for _, srcCM := range configMapList.Items {
		newCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      srcCM.Name,
				Namespace: targetNamespace,
				Labels:    srcCM.Labels,
			},
			Data: srcCM.Data,
		}

		// Update the environment label
		if newCM.Labels == nil {
			newCM.Labels = make(map[string]string)
		}
		newCM.Labels["kloudlite.io/environment"] = environment.Name

		if err := r.Create(ctx, newCM); err != nil && !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to clone configmap",
				zap.String("name", srcCM.Name),
				zap.Error(err))
			// Continue with other resources instead of failing completely
			continue
		}
		clonedConfigMaps++
		logger.Debug("Cloned configmap", zap.String("name", srcCM.Name))
	}
	logger.Info("ConfigMap cloning completed", zap.Int("cloned", clonedConfigMaps), zap.Int("total", len(configMapList.Items)))

	// Clone Secrets with label "kloudlite.io/resource-type: environment-config"
	logger.Info("Cloning Secrets from source environment")
	secretList := &corev1.SecretList{}
	err = r.List(ctx, secretList,
		client.InNamespace(sourceNamespace),
		client.MatchingLabels{"kloudlite.io/resource-type": "environment-config"})
	if err != nil {
		logger.Error("Failed to list source secrets", zap.Error(err))
		if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateError,
			fmt.Sprintf("Failed to list Secrets from source environment '%s': %v", sourceName, err), logger); err != nil {
			logger.Error("Failed to update environment status after retries", zap.Error(err))
		}
		return reconcile.Result{}, fmt.Errorf("failed to list source Secrets: %w", err)
	}

	clonedSecrets := 0
	for _, srcSecret := range secretList.Items {
		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      srcSecret.Name,
				Namespace: targetNamespace,
				Labels:    srcSecret.Labels,
			},
			Type: srcSecret.Type,
			Data: srcSecret.Data,
		}

		// Update the environment label
		if newSecret.Labels == nil {
			newSecret.Labels = make(map[string]string)
		}
		newSecret.Labels["kloudlite.io/environment"] = environment.Name

		if err := r.Create(ctx, newSecret); err != nil && !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to clone secret",
				zap.String("name", srcSecret.Name),
				zap.Error(err))
			// Continue with other resources instead of failing completely
			continue
		}
		clonedSecrets++
		logger.Debug("Cloned secret", zap.String("name", srcSecret.Name))
	}
	logger.Info("Secret cloning completed", zap.Int("cloned", clonedSecrets), zap.Int("total", len(secretList.Items)))

	// Transition to PVC cloning phase after completing resource cloning
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseCloningResources {
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseCloningPVCs, "Starting PVC cloning", logger)
		return reconcile.Result{Requeue: true}, nil
	}

	// Phase 3: Clone PVCs (create empty PVCs in target namespace)
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseCloningPVCs {
		return r.handleCloningPVCsPhase(ctx, environment, sourceNamespace, targetNamespace, logger)
	}

	// Phase 4: Create ALL copy jobs (sender + receiver)
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseCreatingCopyJobs {
		return r.handleCreatingCopyJobsPhase(ctx, environment, sourceName, sourceNamespace, targetNamespace, logger)
	}

	// Phase 5: Wait for ALL copy jobs to complete
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseWaitingForCopyCompletion {
		return r.handleWaitingForCopyCompletionPhase(ctx, environment, sourceNamespace, targetNamespace, logger)
	}

	// Phase 6: Verify copies and cleanup
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseVerifyingCopies {
		return r.handleVerifyingCopiesPhase(ctx, environment, sourceNamespace, targetNamespace, logger)
	}

	// Phase 6.5: Clone Compositions (after PVC data is copied)
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseCloningCompositions {
		return r.handleCloningCompositionsPhase(ctx, environment, sourceName, logger)
	}

	// Phase 7: Resume source environment
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseResuming {
		return r.handleResumingPhase(ctx, environment, sourceName, logger)
	}

	// Only execute completion logic when phase is Completed
	if environment.Status.CloningStatus.Phase != environmentsv1.CloningPhaseCompleted {
		return reconcile.Result{}, nil
	}

	// Prepare cloning completion message with statistics
	sourceName = environment.Spec.CloneFrom
	successMessage := fmt.Sprintf("Successfully cloned environment from %s", sourceName)

	if environment.Status.CloningStatus.TotalPVCs > 0 {
		successMessage = fmt.Sprintf("Successfully cloned environment from %s (PVCs: %d/%d)",
			sourceName, environment.Status.CloningStatus.ClonedPVCs, environment.Status.CloningStatus.TotalPVCs)
	}

	// Update status to indicate cloning is complete
	desiredState := environmentsv1.EnvironmentStateInactive
	if environment.Spec.Activated {
		desiredState = environmentsv1.EnvironmentStateActive
	}

	if err := r.updateEnvironmentStatus(ctx, environment, desiredState, successMessage, logger); err != nil {
		logger.Error("Failed to update environment status after cloning, even after retries", zap.Error(err))
	}

	// Update status with condition for successful cloning
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		r.addOrUpdateCondition(environment, environmentsv1.EnvironmentConditionCloned, metav1.ConditionTrue, "CloningSuccessful", successMessage)
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update environment conditions after retries", zap.Error(err))
	}

	// Mark completion time
	now := metav1.Now()
	environment.Status.CloningStatus.CompletionTime = &now
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update completion time", zap.Error(err))
	}

	// Clear the CloneFrom field to mark cloning as complete
	environment.Spec.CloneFrom = ""
	if err := r.Update(ctx, environment); err != nil {
		logger.Error("Failed to clear CloneFrom field", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Environment cloning completed successfully",
		zap.String("source", sourceName),
		zap.Int32("clonedPVCs", environment.Status.CloningStatus.ClonedPVCs),
		zap.Int32("totalPVCs", environment.Status.CloningStatus.TotalPVCs))

	return reconcile.Result{Requeue: true}, nil
}

// handleCloningPVCsPhase handles Phase 3: Clone PVCs (create empty PVCs in target namespace)
func (r *EnvironmentReconciler) handleCloningPVCsPhase(ctx context.Context, environment *environmentsv1.Environment, sourceNamespace, targetNamespace string, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Creating PVCs in target namespace")

	// List PVCs from source namespace with kloudlite.io/managed label
	pvcList := &corev1.PersistentVolumeClaimList{}
	err := r.List(ctx, pvcList,
		client.InNamespace(sourceNamespace),
		client.MatchingLabels{"kloudlite.io/managed": "true"})
	if err != nil {
		logger.Error("Failed to list source PVCs", zap.Error(err))
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed, fmt.Sprintf("Failed to list PVCs: %v", err), logger)
		return reconcile.Result{}, fmt.Errorf("failed to list source PVCs: %w", err)
	}

	// Update total PVC count
	environment.Status.CloningStatus.TotalPVCs = int32(len(pvcList.Items))
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update PVC count", zap.Error(err))
	}

	// Create empty PVCs in target namespace (data will be copied later)
	for _, srcPVC := range pvcList.Items {
		newPVC := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      srcPVC.Name,
				Namespace: targetNamespace,
				Labels:    srcPVC.Labels,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes:      srcPVC.Spec.AccessModes,
				StorageClassName: srcPVC.Spec.StorageClassName,
				Resources:        srcPVC.Spec.Resources,
			},
		}

		// Update environment label
		if newPVC.Labels == nil {
			newPVC.Labels = make(map[string]string)
		}
		newPVC.Labels["kloudlite.io/environment"] = environment.Name

		if err := r.Create(ctx, newPVC); err != nil && !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to create PVC",
				zap.String("name", srcPVC.Name),
				zap.Error(err))
			// Track failed PVC
			environment.Status.CloningStatus.FailedPVCs = append(environment.Status.CloningStatus.FailedPVCs, srcPVC.Name)
			continue
		}
		logger.Debug("Created PVC", zap.String("name", srcPVC.Name))
	}

	// Move to creating copy jobs phase
	r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseCreatingCopyJobs, "Creating copy jobs", logger)
	return reconcile.Result{Requeue: true}, nil
}

// handleCreatingCopyJobsPhase handles Phase 4: Create ALL copy jobs (sender + receiver)
func (r *EnvironmentReconciler) handleCreatingCopyJobsPhase(ctx context.Context, environment *environmentsv1.Environment, sourceName, sourceNamespace, targetNamespace string, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Creating all PVC copy jobs")

	// List PVCs from source namespace
	pvcList := &corev1.PersistentVolumeClaimList{}
	err := r.List(ctx, pvcList,
		client.InNamespace(sourceNamespace),
		client.MatchingLabels{"kloudlite.io/managed": "true"})
	if err != nil {
		logger.Error("Failed to list source PVCs", zap.Error(err))
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed, fmt.Sprintf("Failed to list PVCs: %v", err), logger)
		return reconcile.Result{}, fmt.Errorf("failed to list source PVCs: %w", err)
	}

	if len(pvcList.Items) == 0 {
		logger.Info("No PVCs to clone, moving to clone compositions phase")
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseCloningCompositions, "No PVCs to clone, cloning Compositions", logger)
		return reconcile.Result{Requeue: true}, nil
	}

	// Initialize copy jobs status tracking
	copyJobsStatus := make([]environmentsv1.PVCCopyJobStatus, 0, len(pvcList.Items))
	copier := NewPVCCopier(r.Client, sourceNamespace, targetNamespace)
	now := metav1.Now()

	// Create sender and receiver jobs for ALL PVCs
	for _, srcPVC := range pvcList.Items {
		jobStatus := environmentsv1.PVCCopyJobStatus{
			PVCName:         srcPVC.Name,
			Phase:           "Creating",
			SenderJobName:   fmt.Sprintf("pvc-copy-sender-%s", srcPVC.Name),
			ReceiverJobName: fmt.Sprintf("pvc-copy-receiver-%s", srcPVC.Name),
			StartTime:       &now,
		}

		logger.Info("Creating copy jobs for PVC", zap.String("pvc", srcPVC.Name))
		if err := copier.CopyPVC(ctx, srcPVC.Name, srcPVC.Name, environment); err != nil {
			logger.Error("Failed to create copy jobs", zap.String("pvc", srcPVC.Name), zap.Error(err))
			jobStatus.Phase = "Failed"
			jobStatus.ErrorMessage = err.Error()
			environment.Status.CloningStatus.FailedPVCs = append(environment.Status.CloningStatus.FailedPVCs, srcPVC.Name)
		} else {
			jobStatus.Phase = "Copying"
		}
		copyJobsStatus = append(copyJobsStatus, jobStatus)
	}

	// Update status with copy jobs info
	environment.Status.CloningStatus.CopyJobsStatus = copyJobsStatus
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update copy jobs status", zap.Error(err))
	}

	// Update source environment to Copying phase
	sourceEnv := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: sourceName}, sourceEnv); err == nil {
		if sourceEnv.Status.SourceCloningStatus != nil {
			sourceEnv.Status.SourceCloningStatus.Phase = environmentsv1.SourceCloningPhaseCopying
			sourceEnv.Status.SourceCloningStatus.Message = "PVC data is being copied"
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, sourceEnv, func() error {
				return nil
			}, logger); err != nil {
				logger.Error("Failed to update source copying status", zap.Error(err))
			}
		}
	}

	logger.Info("All copy jobs created", zap.Int("total", len(copyJobsStatus)))
	r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseWaitingForCopyCompletion,
		fmt.Sprintf("Waiting for %d PVC copies to complete", len(copyJobsStatus)), logger)
	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
}

// handleWaitingForCopyCompletionPhase handles Phase 5: Wait for ALL copy jobs to complete
func (r *EnvironmentReconciler) handleWaitingForCopyCompletionPhase(ctx context.Context, environment *environmentsv1.Environment, sourceNamespace, targetNamespace string, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Checking status of all PVC copy jobs")

	copier := NewPVCCopier(r.Client, sourceNamespace, targetNamespace)
	allCompleted := true
	anyFailed := false
	completedCount := int32(0)
	now := metav1.Now()

	// Check status of ALL copy jobs
	for i := range environment.Status.CloningStatus.CopyJobsStatus {
		jobStatus := &environment.Status.CloningStatus.CopyJobsStatus[i]

		if jobStatus.Phase == "Completed" {
			completedCount++
			continue
		}

		if jobStatus.Phase == "Failed" {
			anyFailed = true
			continue
		}

		// Check receiver job status
		completed, failed, err := copier.GetCopyStatus(ctx, jobStatus.PVCName)
		if err != nil {
			logger.Warn("Failed to check copy status", zap.String("pvc", jobStatus.PVCName), zap.Error(err))
			allCompleted = false
			continue
		}

		if completed {
			jobStatus.Phase = "Completed"
			jobStatus.CompletionTime = &now
			completedCount++
			logger.Info("PVC copy completed", zap.String("pvc", jobStatus.PVCName))
		} else if failed {
			jobStatus.Phase = "Failed"
			jobStatus.ErrorMessage = "Copy job failed"
			anyFailed = true
			environment.Status.CloningStatus.FailedPVCs = append(environment.Status.CloningStatus.FailedPVCs, jobStatus.PVCName)
			logger.Error("PVC copy failed", zap.String("pvc", jobStatus.PVCName))
		} else {
			allCompleted = false
			jobStatus.Phase = "Copying"
		}
	}

	// Update cloned PVCs count
	environment.Status.CloningStatus.ClonedPVCs = completedCount

	// Update status
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update copy status", zap.Error(err))
	}

	if anyFailed && allCompleted {
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed,
			fmt.Sprintf("Some PVC copies failed: %v", environment.Status.CloningStatus.FailedPVCs), logger)
		return reconcile.Result{}, fmt.Errorf("some PVC copies failed")
	}

	if !allCompleted {
		logger.Info("Waiting for PVC copies to complete",
			zap.Int32("completed", completedCount),
			zap.Int32("total", environment.Status.CloningStatus.TotalPVCs))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// ALL copies completed successfully!
	logger.Info("All PVC copies completed successfully",
		zap.Int32("total", completedCount))
	r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseVerifyingCopies, "Verifying copies", logger)
	return reconcile.Result{Requeue: true}, nil
}

// handleVerifyingCopiesPhase handles Phase 6: Verify copies and cleanup
func (r *EnvironmentReconciler) handleVerifyingCopiesPhase(ctx context.Context, environment *environmentsv1.Environment, sourceNamespace, targetNamespace string, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Verifying all PVC copies")

	// Cleanup completed copy jobs
	copier := NewPVCCopier(r.Client, sourceNamespace, targetNamespace)
	for _, jobStatus := range environment.Status.CloningStatus.CopyJobsStatus {
		if jobStatus.Phase == "Completed" {
			if err := copier.CleanupCopyJobs(ctx, jobStatus.PVCName); err != nil {
				logger.Warn("Failed to cleanup copy jobs",
					zap.String("pvc", jobStatus.PVCName),
					zap.Error(err))
			}
		}
	}

	logger.Info("Verification complete, proceeding to clone Compositions")
	r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseCloningCompositions, "Cloning Compositions to destination", logger)
	return reconcile.Result{Requeue: true}, nil
}

// handleCloningCompositionsPhase handles Phase 6.5: Clone Compositions (after PVC data is copied)
func (r *EnvironmentReconciler) handleCloningCompositionsPhase(ctx context.Context, environment *environmentsv1.Environment, sourceName string, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Cloning Compositions from source environment to destination (after PVC data is ready)")

	sourceNamespace := fmt.Sprintf("env-%s", sourceName)
	targetNamespace := environment.Spec.TargetNamespace

	// Clone Compositions
	compositionList := &environmentsv1.CompositionList{}
	err := r.List(ctx, compositionList, client.InNamespace(sourceNamespace))
	if err != nil {
		logger.Error("Failed to list source compositions", zap.Error(err))
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed,
			fmt.Sprintf("Failed to list Compositions from source environment '%s': %v", sourceName, err), logger)
		return reconcile.Result{}, fmt.Errorf("failed to list source Compositions: %w", err)
	}

	clonedCompositions := 0
	for _, srcComp := range compositionList.Items {
		// Deep copy the spec to avoid sharing map/slice references
		newSpec := srcComp.Spec.DeepCopy()

		// Override NodeName with target environment's NodeName
		// This ensures composition pods run on the correct node where target PVCs are bound
		newSpec.NodeName = environment.Spec.NodeName

		// Create new labels map
		newLabels := make(map[string]string)
		for k, v := range srcComp.Labels {
			newLabels[k] = v
		}
		newLabels["kloudlite.io/environment"] = environment.Name

		// Create new annotations map
		newAnnotations := make(map[string]string)
		for k, v := range srcComp.Annotations {
			newAnnotations[k] = v
		}

		newComp := &environmentsv1.Composition{
			ObjectMeta: metav1.ObjectMeta{
				Name:        srcComp.Name,
				Namespace:   targetNamespace,
				Labels:      newLabels,
				Annotations: newAnnotations,
			},
			Spec: *newSpec,
		}

		if err := r.Create(ctx, newComp); err != nil && !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to clone composition",
				zap.String("name", srcComp.Name),
				zap.Error(err))
			// Continue with other resources instead of failing completely
			continue
		}
		clonedCompositions++
		logger.Debug("Cloned composition", zap.String("name", srcComp.Name))
	}
	logger.Info("Composition cloning completed", zap.Int("cloned", clonedCompositions), zap.Int("total", len(compositionList.Items)))

	// Transition to Resuming phase
	r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseResuming, "Resuming source environment", logger)
	return reconcile.Result{Requeue: true}, nil
}

// handleResumingPhase handles Phase 7: Resume source environment
func (r *EnvironmentReconciler) handleResumingPhase(ctx context.Context, environment *environmentsv1.Environment, sourceName string, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Resuming source environment")

	// Reload source environment
	sourceEnv := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: sourceName}, sourceEnv); err != nil {
		logger.Error("Failed to get source environment for resuming", zap.Error(err))
	} else {
		// Update source cloning status to Resuming
		if sourceEnv.Status.SourceCloningStatus != nil {
			sourceEnv.Status.SourceCloningStatus.Phase = environmentsv1.SourceCloningPhaseResuming
			sourceEnv.Status.SourceCloningStatus.Message = "Resuming after cloning completed"
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, sourceEnv, func() error {
				return nil
			}, logger); err != nil {
				logger.Error("Failed to update source resuming status", zap.Error(err))
			}
		}

		// Scale up deployments back to original replica counts
		if err := r.resumeEnvironment(ctx, sourceEnv, logger); err != nil {
			logger.Error("Failed to resume source environment", zap.Error(err))
		} else {
			logger.Info("Source environment resumed successfully")

			// Clear source cloning status after successful resume
			if sourceEnv.Status.SourceCloningStatus != nil {
				sourceEnv.Status.SourceCloningStatus = nil
				if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, sourceEnv, func() error {
					return nil
				}, logger); err != nil {
					logger.Error("Failed to clear source cloning status", zap.Error(err))
				}
			}
		}
	}

	// Move to completed phase
	r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseCompleted, "Cloning completed successfully", logger)
	return reconcile.Result{Requeue: true}, nil
}
